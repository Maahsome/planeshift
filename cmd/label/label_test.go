package label

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"planeshift/config"
	"planeshift/help"
	"planeshift/plane"
)

func TestLabelCommandRegistrationAndFlags(t *testing.T) {
	command := Init(&config.Config{}, nil)
	if command.Name() != config.LabelCommandName || !command.HasAlias(config.LabelCommandAlias) {
		t.Fatalf("command = %q aliases %v", command.Name(), command.Aliases)
	}
	wantArgs := map[string]int{"list": 0, "create": 0, "get": 1, "update": 1, "delete": 1}
	for name, count := range wantArgs {
		child, _, err := command.Find([]string{name})
		if err != nil || child == nil {
			t.Fatalf("find %s: %v", name, err)
		}
		if err := child.Args(child, make([]string, count+1)); err == nil {
			t.Fatalf("%s accepted too many args", name)
		}
		if err := child.Args(child, make([]string, count)); err != nil {
			t.Fatalf("%s rejected exact args: %v", name, err)
		}
	}
	for _, name := range []string{"list", "create", "get", "update", "delete"} {
		child, _, _ := command.Find([]string{name})
		for _, flag := range []string{"workspace", "project-id"} {
			if child.Flags().Lookup(flag) == nil {
				t.Fatalf("%s missing --%s", name, flag)
			}
		}
	}
	list, _, _ := command.Find([]string{"list"})
	for _, flag := range []string{"cursor", "per-page", "fields", "expand", "order-by"} {
		if list.Flags().Lookup(flag) == nil {
			t.Fatalf("list missing --%s", flag)
		}
	}
	for _, flag := range []string{"name", "color", "description", "external-source", "external-id", "parent", "sort-order"} {
		if list.Flags().Lookup(flag) != nil {
			t.Fatalf("list exposed unsupported --%s", flag)
		}
	}
	create, _, _ := command.Find([]string{"create"})
	for _, flag := range []string{"name", "color", "description", "external-source", "external-id", "parent", "sort-order"} {
		if create.Flags().Lookup(flag) == nil {
			t.Fatalf("create missing --%s", flag)
		}
	}
	update, _, _ := command.Find([]string{"update"})
	for _, flag := range []string{"name", "color", "description", "external-source", "external-id", "parent", "sort-order"} {
		if update.Flags().Lookup(flag) == nil {
			t.Fatalf("update missing --%s", flag)
		}
	}
	if command.Flags().Lookup("api-key") != nil {
		t.Fatal("label command exposed credential flags")
	}
}

func TestLabelCommandAliasAndLazyFactoryUseCentralOutput(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeLabelClient{response: `{"id":"label-1","name":"Urgent","future":{"keep":true}}`}
	var calls int
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "team", Project: config.ProjectContext{ID: "project"}}}
	clientFactory = func() (plane.Client, error) { calls++; return fake, nil }
	command := newGetCommand()
	if calls != 0 {
		t.Fatal("factory invoked during command construction")
	}
	command.SetContext(context.Background())
	output := captureLabelStdout(t, func() {
		if err := runGet(command, []string{"label"}); err != nil {
			t.Fatal(err)
		}
	})
	if calls != 1 || len(fake.routes) != 1 {
		t.Fatalf("factory/routes = %d/%v", calls, fake.routes)
	}
	if fake.routes[0] != "/workspaces/team/projects/project/labels/label/" {
		t.Fatalf("label route = %q", fake.routes[0])
	}
	if !strings.Contains(output, `"future":{"keep":true}`) {
		t.Fatalf("raw output lost unknown field: %s", output)
	}
}

func TestLabelContextDefaultsOverridesAndFailures(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeLabelClient{response: `{"id":"label-1","name":"Urgent"}`}
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "saved-workspace", Project: config.ProjectContext{ID: "saved-project"}}}
	var calls int
	clientFactory = func() (plane.Client, error) { calls++; return fake, nil }

	defaultCommand := newGetCommand()
	if err := runGet(defaultCommand, []string{"label"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[0] != "/workspaces/saved-workspace/projects/saved-project/labels/label/" {
		t.Fatalf("context default route = %q", fake.routes[0])
	}

	override := newGetCommand()
	if err := override.Flags().Set("workspace", "override-workspace"); err != nil {
		t.Fatal(err)
	}
	if err := override.Flags().Set("project-id", "override-project"); err != nil {
		t.Fatal(err)
	}
	if err := runGet(override, []string{"label"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[1] != "/workspaces/override-workspace/projects/override-project/labels/label/" {
		t.Fatalf("override route = %q", fake.routes[1])
	}
	if c.Context.Workspace != "saved-workspace" || c.Context.Project.ID != "saved-project" {
		t.Fatalf("command override changed context: %#v", c.Context)
	}

	blank := newGetCommand()
	if err := blank.Flags().Set("project-id", ""); err != nil {
		t.Fatal(err)
	}
	if err := runGet(blank, []string{"label"}); err == nil || calls != 2 {
		t.Fatalf("blank override result=%v factory calls=%d", err, calls)
	}

	c.Context = config.Context{Workspace: "saved-workspace"}
	missing := newGetCommand()
	if err := runGet(missing, []string{"label"}); err == nil || calls != 2 {
		t.Fatalf("missing project result=%v factory calls=%d", err, calls)
	}
}

func TestLabelCreateAndUpdatePreserveChangedValues(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeLabelClient{response: `{"id":"label-1","name":"Urgent"}`}
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "team", Project: config.ProjectContext{ID: "project"}}}
	clientFactory = func() (plane.Client, error) { return fake, nil }

	create := newCreateCommand()
	for name, value := range map[string]string{
		"name": "Urgent", "color": "", "description": "", "external-source": "github",
		"external-id": "", "parent": "", "sort-order": "0",
	} {
		if err := create.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	create.SetContext(context.Background())
	if err := runCreate(create, nil); err != nil {
		t.Fatal(err)
	}
	assertLabelBodyFields(t, fake.body, []string{"name", "color", "description", "external_source", "external_id", "parent", "sort_order"})
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(fake.body, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["color"]) != `""` || string(fields["sort_order"]) != "0" {
		t.Fatalf("create presence values = %s", fake.body)
	}

	fake.body = nil
	update := newUpdateCommand()
	for name, value := range map[string]string{"description": "", "parent": "", "sort-order": "0"} {
		if err := update.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	update.SetContext(context.Background())
	if err := runUpdate(update, []string{"label"}); err != nil {
		t.Fatal(err)
	}
	assertLabelBodyFields(t, fake.body, []string{"description", "parent", "sort_order"})
	if string(fake.body) != `{"description":"","parent":"","sort_order":0}` {
		t.Fatalf("update body = %s", fake.body)
	}

	fake.body = nil
	emptyUpdate := newUpdateCommand()
	if err := runUpdate(emptyUpdate, []string{"label"}); err != nil {
		t.Fatal(err)
	}
	if string(fake.body) != `{}` {
		t.Fatalf("empty update body = %s", fake.body)
	}
}

func TestLabelListBuildsOnlyDocumentedQueryAndValidatesBeforeFactory(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeLabelClient{response: `{"results":[]}`}
	var calls int
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "team", Project: config.ProjectContext{ID: "project"}}}
	clientFactory = func() (plane.Client, error) { calls++; return fake, nil }
	invalid := newListCommand()
	if err := invalid.Flags().Set("per-page", "101"); err != nil {
		t.Fatal(err)
	}
	invalid.SetContext(context.Background())
	if err := runList(invalid, nil); err == nil {
		t.Fatal("list accepted invalid page size")
	}
	if calls != 0 {
		t.Fatal("list resolved factory before validating page size")
	}

	list := newListCommand()
	for name, value := range map[string]string{
		"cursor": "20:1:0", "per-page": "20", "fields": "id,name", "expand": "parent", "order-by": "-sort_order",
	} {
		if err := list.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	list.SetContext(context.Background())
	if err := runList(list, nil); err != nil {
		t.Fatal(err)
	}
	if len(fake.queries) != 1 || fake.queries[0].Encode() != "cursor=20%3A1%3A0&expand=parent&fields=id%2Cname&order_by=-sort_order&per_page=20" {
		t.Fatalf("list query = %v", fake.queries)
	}
}

func TestLabelDeleteIsQuietAndHelpIsSafe(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeLabelClient{status: http.StatusNoContent}
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "team", Project: config.ProjectContext{ID: "project"}}}
	clientFactory = func() (plane.Client, error) { return fake, nil }
	command := newDeleteCommand()
	command.SetContext(context.Background())
	output := captureLabelStdout(t, func() {
		if err := runDelete(command, []string{"label"}); err != nil {
			t.Fatal(err)
		}
	})
	if output != "" {
		t.Fatalf("delete output = %q, want empty", output)
	}
	helpText := strings.ToLower((&help.LabelCmd{}).Long())
	for _, phrase := range []string{
		"label", "labels", "context.workspace", "context.project.id", "list", "create", "get", "update", "delete",
		"cursor", "per-page", "fields", "expand", "order-by", "name", "color", "description", "external-source",
		"external-id", "parent", "sort-order", "201", "200", "204", "json", "yaml", "gron", "text", "table", "raw",
	} {
		if !strings.Contains(helpText, phrase) {
			t.Fatalf("help omitted %q", phrase)
		}
	}
	for _, secret := range []string{"api-key", "bearer", "presigned", "invitation", "upload", "token"} {
		if strings.Contains(helpText, secret) {
			t.Fatalf("help contains prohibited %q", secret)
		}
	}
	rootHelp := strings.ToLower((&help.RootCmd{}).Long())
	if !strings.Contains(rootHelp, "label") || !strings.Contains(rootHelp, "labels") {
		t.Fatalf("root help omitted label discoverability: %s", rootHelp)
	}
}

type fakeLabelClient struct {
	routes   []string
	queries  []url.Values
	body     []byte
	response string
	status   int
}

func (f *fakeLabelClient) Do(_ context.Context, _ string, route string, query url.Values, body any, _ http.Header, destination any) (plane.Response, error) {
	f.routes = append(f.routes, route)
	f.queries = append(f.queries, query)
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return plane.Response{}, err
		}
		f.body = data
	}
	if destination != nil && f.response != "" {
		if err := json.Unmarshal([]byte(f.response), destination); err != nil {
			return plane.Response{}, err
		}
	}
	status := f.status
	if status == 0 {
		status = http.StatusOK
	}
	return plane.Response{StatusCode: status}, nil
}

func assertLabelBodyFields(t *testing.T, body []byte, want []string) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range want {
		if _, ok := fields[key]; !ok {
			t.Fatalf("body omitted %q: %s", key, body)
		}
	}
}

func captureLabelStdout(t *testing.T, function func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = writer
	function()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = previous
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	_ = reader.Close()
	return string(data)
}

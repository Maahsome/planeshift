package state

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

func TestStateCommandRegistrationAndFlags(t *testing.T) {
	command := Init(&config.Config{}, nil)
	if command.Name() != config.StateCommandName || !command.HasAlias(config.StateCommandAlias) {
		t.Fatalf("command = %q aliases %v", command.Name(), command.Aliases)
	}
	wantArgs := map[string]int{"list": 2, "create": 2, "get": 3, "update": 3, "delete": 3}
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
	list, _, _ := command.Find([]string{"list"})
	for _, flag := range []string{"cursor", "per-page", "fields", "expand"} {
		if list.Flags().Lookup(flag) == nil {
			t.Fatalf("list missing --%s", flag)
		}
	}
	if list.Flags().Lookup("order-by") != nil {
		t.Fatal("list exposed unsupported --order-by")
	}
	create, _, _ := command.Find([]string{"create"})
	for _, flag := range []string{"name", "color", "description", "sequence", "group", "is-triage", "default", "external-source", "external-id"} {
		if create.Flags().Lookup(flag) == nil {
			t.Fatalf("create missing --%s", flag)
		}
	}
	update, _, _ := command.Find([]string{"update"})
	if update.Flags().Lookup("name") == nil || update.Flags().Lookup("color") == nil || update.Flags().Lookup("sequence") == nil {
		t.Fatalf("update is missing documented fields")
	}
	if command.Flags().Lookup("api-key") != nil {
		t.Fatal("state command exposed credential flags")
	}
}

func TestStateCommandsResolveFactoryLazilyAndUseCentralOutput(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeStateClient{response: `{"id":"state-1","name":"Started","future":{"keep":true}}`}
	var calls int
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) { calls++; return fake, nil }
	command := newGetCommand()
	if calls != 0 {
		t.Fatal("factory invoked during command construction")
	}
	command.SetContext(context.Background())
	output := captureStateStdout(t, func() {
		if err := runGet(command, []string{"team", "project", "state"}); err != nil {
			t.Fatal(err)
		}
	})
	if calls != 1 || len(fake.routes) != 1 {
		t.Fatalf("factory/routes = %d/%v", calls, fake.routes)
	}
	if fake.routes[0] != "/workspaces/team/projects/project/states/state/" {
		t.Fatalf("state route = %q", fake.routes[0])
	}
	if !strings.Contains(output, `"future":{"keep":true}`) {
		t.Fatalf("raw output lost unknown field: %s", output)
	}
}

func TestStateCreateAndUpdatePreserveChangedValues(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeStateClient{response: `{"id":"state-1","name":"Started"}`}
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) { return fake, nil }

	create := newCreateCommand()
	for name, value := range map[string]string{
		"name": "Started", "color": "#00ff00", "description": "", "sequence": "null",
		"group": "triage", "is-triage": "false", "default": "false", "external-id": "",
	} {
		if err := create.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	create.SetContext(context.Background())
	if err := runCreate(create, []string{"team", "project"}); err != nil {
		t.Fatal(err)
	}
	assertBodyFields(t, fake.body, []string{"name", "color", "description", "sequence", "group", "is_triage", "default", "external_id"})

	fake.body = nil
	update := newUpdateCommand()
	for name, value := range map[string]string{"description": "", "is-triage": "false", "sequence": "0"} {
		if err := update.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	update.SetContext(context.Background())
	if err := runUpdate(update, []string{"team", "project", "state"}); err != nil {
		t.Fatal(err)
	}
	assertBodyFields(t, fake.body, []string{"description", "is_triage", "sequence"})
	var updateFields map[string]json.RawMessage
	if err := json.Unmarshal(fake.body, &updateFields); err != nil {
		t.Fatal(err)
	}
	if string(updateFields["description"]) != `""` || string(updateFields["is_triage"]) != "false" || string(updateFields["sequence"]) != "0" {
		t.Fatalf("update presence values = %s", fake.body)
	}
}

func TestStateListBuildsOnlyDocumentedQueryAndValidatesBeforeFactory(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeStateClient{response: `{"results":[]}`}
	var calls int
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) { calls++; return fake, nil }
	invalid := newListCommand()
	if err := invalid.Flags().Set("per-page", "101"); err != nil {
		t.Fatal(err)
	}
	invalid.SetContext(context.Background())
	if err := runList(invalid, []string{"team", "project"}); err == nil {
		t.Fatal("list accepted invalid page size")
	}
	if calls != 0 {
		t.Fatal("list resolved factory before validating page size")
	}

	list := newListCommand()
	for name, value := range map[string]string{"cursor": "20:1:0", "per-page": "20", "fields": "id,name", "expand": "project"} {
		if err := list.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	list.SetContext(context.Background())
	if err := runList(list, []string{"team", "project"}); err != nil {
		t.Fatal(err)
	}
	if len(fake.queries) != 1 || fake.queries[0].Encode() != "cursor=20%3A1%3A0&expand=project&fields=id%2Cname&per_page=20" {
		t.Fatalf("list query = %v", fake.queries)
	}
}

func TestStateDeleteIsQuietAndHelpIsSafe(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeStateClient{status: http.StatusNoContent}
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) { return fake, nil }
	command := newDeleteCommand()
	command.SetContext(context.Background())
	output := captureStateStdout(t, func() {
		if err := runDelete(command, []string{"team", "project", "state"}); err != nil {
			t.Fatal(err)
		}
	})
	if output != "" {
		t.Fatalf("delete output = %q, want empty", output)
	}
	helpText := strings.ToLower((&help.StateCmd{}).Long())
	for _, phrase := range []string{"workspace slug", "project id", "list", "create", "update", "delete", "sequence", "204"} {
		if !strings.Contains(helpText, phrase) {
			t.Fatalf("help omitted %q", phrase)
		}
	}
	for _, secret := range []string{"api-key", "bearer", "presigned", "invitation", "upload", "token"} {
		if strings.Contains(helpText, secret) {
			t.Fatalf("help contains prohibited %q", secret)
		}
	}
}

type fakeStateClient struct {
	routes   []string
	queries  []url.Values
	body     []byte
	response string
	status   int
}

func (f *fakeStateClient) Do(_ context.Context, method, route string, query url.Values, body any, _ http.Header, destination any) (plane.Response, error) {
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

func assertBodyFields(t *testing.T, body []byte, want []string) {
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

func captureStateStdout(t *testing.T, function func()) string {
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

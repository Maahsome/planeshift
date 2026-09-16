package workitem

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
	"planeshift/plane"
)

func TestWorkItemCommandRegistrationAndCompatibilityBoundary(t *testing.T) {
	command := Init(&config.Config{}, nil)
	if command.Name() != config.WorkItemCommandName || !command.HasAlias(config.WorkItemCommandAlias) {
		t.Fatalf("command = %q aliases %v", command.Name(), command.Aliases)
	}
	wantArgs := map[string]int{
		"search": 1, "get-by-identifier": 3, "list": 2, "create": 2, "get": 3,
		"update": 3, "delete": 3, "relations-list": 3, "relations-create": 3,
	}
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
	for _, flag := range []string{"cursor", "per-page", "fields", "expand", "external-id", "external-source", "order-by"} {
		if list.Flags().Lookup(flag) == nil {
			t.Fatalf("list missing --%s", flag)
		}
	}
	search, _, _ := command.Find([]string{"search"})
	for _, flag := range []string{"search", "limit", "project-id", "workspace-search"} {
		if search.Flags().Lookup(flag) == nil {
			t.Fatalf("search missing --%s", flag)
		}
	}
	relationsCreate, _, _ := command.Find([]string{"relations-create"})
	for _, flag := range []string{"relation-type", "issue"} {
		if relationsCreate.Flags().Lookup(flag) == nil {
			t.Fatalf("relations-create missing --%s", flag)
		}
	}
	legacy, _, err := command.Find([]string{"legacy"})
	if err != nil || legacy == nil || !legacy.Hidden {
		t.Fatalf("legacy command = %v/%v hidden=%v", legacy, err, legacy != nil && legacy.Hidden)
	}
	if relations, _, err := legacy.Find([]string{"relations-list"}); err == nil && relations != nil && relations != legacy {
		t.Fatalf("legacy unexpectedly exposes relations-list")
	}
	if issue, _, err := command.Find([]string{"issue"}); err == nil && issue != nil && issue != command {
		t.Fatalf("competing issue command exists: %v", issue)
	}
}

func TestWorkItemCommandsResolveFactoryLazilyAndSelectRoutes(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeWorkItemClient{response: `{"id":"item-1","name":"Item","future":{"keep":true}}`}
	var calls int
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) { calls++; return fake, nil }
	command := newGetCommand(false)
	command.SetContext(context.Background())
	if calls != 0 {
		t.Fatal("factory invoked during command construction")
	}
	output := captureWorkItemStdout(t, func() {
		if err := runGet(command, []string{"team", "project", "item"}); err != nil {
			t.Fatal(err)
		}
	})
	if calls != 1 || len(fake.routes) != 1 {
		t.Fatalf("factory/routes = %d/%v", calls, fake.routes)
	}
	if fake.routes[0] != "/workspaces/team/projects/project/work-items/item/" {
		t.Fatalf("current route = %q", fake.routes[0])
	}
	if !strings.Contains(output, `"future":{"keep":true}`) {
		t.Fatalf("raw output lost unknown field: %s", output)
	}

	legacyCommand := newGetCommand(true)
	legacyCommand.SetContext(context.Background())
	if err := runLegacyGet(legacyCommand, []string{"team", "project", "item"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[1] != "/workspaces/team/projects/project/issues/item/" {
		t.Fatalf("legacy route = %q", fake.routes[1])
	}
}

func TestWorkItemUpdatePreservesExplicitFlagValues(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeWorkItemClient{response: `{"id":"item-1","name":"Item"}`}
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) { return fake, nil }
	command := newUpdateCommand(false)
	for name, value := range map[string]string{"is-draft": "false", "point": "0", "assignees": ""} {
		if err := command.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	command.SetContext(context.Background())
	if err := runUpdate(command, []string{"team", "project", "item"}); err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(fake.body, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"is_draft", "point", "assignees"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("request omitted explicit %s: %s", key, fake.body)
		}
	}
}

func TestWorkItemHelpIsSafe(t *testing.T) {
	help := Init(&config.Config{}, nil).Long
	for _, phrase := range []string{"workspace slug", "project ID", "search", "relations", "legacy", "204", "/work-items/"} {
		if !strings.Contains(help, phrase) {
			t.Fatalf("help omitted %q", phrase)
		}
	}
	for _, secret := range []string{"api-key", "bearer", "presigned", "invitation", "upload"} {
		if strings.Contains(strings.ToLower(help), secret) {
			t.Fatalf("help contains prohibited %q", secret)
		}
	}
}

type fakeWorkItemClient struct {
	routes   []string
	body     []byte
	response string
}

func (f *fakeWorkItemClient) Do(ctx context.Context, method, route string, query url.Values, body any, headers http.Header, destination any) (plane.Response, error) {
	f.routes = append(f.routes, route)
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return plane.Response{}, err
		}
		f.body = data
	}
	if destination != nil {
		if err := json.Unmarshal([]byte(f.response), destination); err != nil {
			return plane.Response{}, err
		}
	}
	return plane.Response{StatusCode: http.StatusOK}, nil
}

func captureWorkItemStdout(t *testing.T, function func()) string {
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

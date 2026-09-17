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
		"search": 0, "get-by-identifier": 2, "list": 0, "create": 0, "get": 1,
		"update": 1, "delete": 1, "relations-list": 1, "relations-create": 1,
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
	for _, flag := range []string{"workspace", "project-id"} {
		if list.Flags().Lookup(flag) == nil {
			t.Fatalf("list missing --%s", flag)
		}
	}
	for _, flag := range []string{"cursor", "per-page", "fields", "expand", "external-id", "external-source", "order-by"} {
		if list.Flags().Lookup(flag) == nil {
			t.Fatalf("list missing --%s", flag)
		}
	}
	search, _, _ := command.Find([]string{"search"})
	for _, flag := range []string{"search", "limit", "project-id", "workspace-search", "workspace"} {
		if search.Flags().Lookup(flag) == nil {
			t.Fatalf("search missing --%s", flag)
		}
	}
	identifier, _, _ := command.Find([]string{"get-by-identifier"})
	if identifier.Flags().Lookup("workspace") == nil {
		t.Fatal("get-by-identifier missing --workspace")
	}
	for _, name := range []string{"create", "get", "update", "delete", "relations-list", "relations-create"} {
		child, _, _ := command.Find([]string{name})
		for _, flag := range []string{"workspace", "project-id"} {
			if child.Flags().Lookup(flag) == nil {
				t.Fatalf("%s missing --%s", name, flag)
			}
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
	legacyArgs := map[string]int{"search": 0, "get-by-identifier": 2, "list": 0, "create": 0, "get": 1, "update": 1, "delete": 1}
	if len(legacy.Commands()) != len(legacyArgs) {
		t.Fatalf("legacy command count = %d, want %d", len(legacy.Commands()), len(legacyArgs))
	}
	for name, count := range legacyArgs {
		child, _, err := legacy.Find([]string{name})
		if err != nil || child == nil {
			t.Fatalf("find legacy %s: %v", name, err)
		}
		if err := child.Args(child, make([]string, count+1)); err == nil {
			t.Fatalf("legacy %s accepted too many args", name)
		}
		if err := child.Args(child, make([]string, count)); err != nil {
			t.Fatalf("legacy %s rejected residual args: %v", name, err)
		}
	}
	if relations, _, err := legacy.Find([]string{"relations-list"}); err == nil && relations != nil && relations != legacy {
		t.Fatalf("legacy unexpectedly exposes relations-list")
	}
	for _, name := range []string{"list", "create", "get", "update", "delete"} {
		child, _, _ := legacy.Find([]string{name})
		for _, flag := range []string{"workspace", "project-id"} {
			if child.Flags().Lookup(flag) == nil {
				t.Fatalf("legacy %s missing --%s", name, flag)
			}
		}
	}
	for _, name := range []string{"search", "get-by-identifier"} {
		child, _, _ := legacy.Find([]string{name})
		if child.Flags().Lookup("workspace") == nil {
			t.Fatalf("legacy %s missing --workspace", name)
		}
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
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "team", Project: config.ProjectContext{ID: "project"}}}
	clientFactory = func() (plane.Client, error) { calls++; return fake, nil }
	command := newGetCommand(false)
	command.SetContext(context.Background())
	if calls != 0 {
		t.Fatal("factory invoked during command construction")
	}
	output := captureWorkItemStdout(t, func() {
		if err := runGet(command, []string{"item"}); err != nil {
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
	if err := runLegacyGet(legacyCommand, []string{"item"}); err != nil {
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
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "team", Project: config.ProjectContext{ID: "project"}}}
	clientFactory = func() (plane.Client, error) { return fake, nil }
	command := newUpdateCommand(false)
	for name, value := range map[string]string{"is-draft": "false", "point": "0", "assignees": ""} {
		if err := command.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	command.SetContext(context.Background())
	if err := runUpdate(command, []string{"item"}); err != nil {
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

func TestRelationsCreateOutputsFlatJSONResponse(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeWorkItemClient{response: `[{"id":"relation-1","relation_type":"relates_to","future":{"keep":true}}]`}
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "team", Project: config.ProjectContext{ID: "project"}}}
	clientFactory = func() (plane.Client, error) { return fake, nil }
	command := newRelationsCreateCommand()
	if err := command.Flags().Set("relation-type", "relates_to"); err != nil {
		t.Fatal(err)
	}
	if err := command.Flags().Set("issue", "issue-1"); err != nil {
		t.Fatal(err)
	}
	command.SetContext(context.Background())
	output := captureWorkItemStdout(t, func() {
		if err := runRelationsCreate(command, []string{"item"}); err != nil {
			t.Fatal(err)
		}
	})
	if len(fake.routes) != 1 || fake.routes[0] != "/workspaces/team/projects/project/work-items/item/relations/" {
		t.Fatalf("relation route = %v", fake.routes)
	}
	var relations []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &relations); err != nil {
		t.Fatalf("output = %q: %v", output, err)
	}
	if len(relations) != 1 || string(relations[0]["id"]) != `"relation-1"` || string(relations[0]["relation_type"]) != `"relates_to"` {
		t.Fatalf("relation output = %s", output)
	}
}

func TestWorkItemContextDefaultsOverridesAndIdentifierExceptions(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeWorkItemClient{response: `{"id":"item-1","name":"Item"}`}
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "saved-workspace", Project: config.ProjectContext{ID: "saved-project"}}}
	var calls int
	clientFactory = func() (plane.Client, error) { calls++; return fake, nil }

	search := newSearchCommand(false)
	_ = search.Flags().Set("search", "release")
	_ = search.Flags().Set("project-id", "query-project")
	if err := runSearch(search, nil); err != nil {
		t.Fatal(err)
	}
	if fake.routes[0] != "/workspaces/saved-workspace/work-items/search/" || fake.queries[0].Get("project_id") != "query-project" {
		t.Fatalf("search route/query = %q/%v", fake.routes[0], fake.queries[0])
	}

	identifier := newGetByIdentifierCommand(false)
	if err := runGetByIdentifier(identifier, []string{"ENG", "123"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[1] != "/workspaces/saved-workspace/work-items/ENG-123/" {
		t.Fatalf("identifier route = %q", fake.routes[1])
	}

	get := newGetCommand(false)
	_ = get.Flags().Set("workspace", "override-workspace")
	_ = get.Flags().Set("project-id", "override-project")
	if err := runGet(get, []string{"item"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[2] != "/workspaces/override-workspace/projects/override-project/work-items/item/" {
		t.Fatalf("override route = %q", fake.routes[2])
	}
	if c.Context.Workspace != "saved-workspace" || c.Context.Project.ID != "saved-project" {
		t.Fatalf("command override changed context: %#v", c.Context)
	}

	blank := newGetCommand(false)
	_ = blank.Flags().Set("workspace", "")
	if err := runGet(blank, []string{"item"}); err == nil || calls != 3 {
		t.Fatalf("blank override result=%v factory calls=%d", err, calls)
	}

	c.Context = config.Context{}
	missing := newGetCommand(false)
	if err := runGet(missing, []string{"item"}); err == nil || calls != 3 {
		t.Fatalf("missing context result=%v factory calls=%d", err, calls)
	}
}

func TestWorkItemHelpIsSafe(t *testing.T) {
	help := Init(&config.Config{}, nil).Long
	for _, phrase := range []string{"context.workspace", "context.project.id", "search", "relations", "legacy", "204", "/work-items/"} {
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
	queries  []url.Values
	body     []byte
	response string
}

func (f *fakeWorkItemClient) Do(ctx context.Context, method, route string, query url.Values, body any, headers http.Header, destination any) (plane.Response, error) {
	f.routes = append(f.routes, route)
	f.queries = append(f.queries, query)
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

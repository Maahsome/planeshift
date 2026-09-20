package activity

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

func TestActivityCommandRegistrationFlagsAndHiddenLegacy(t *testing.T) {
	command := Init(&config.Config{}, nil)
	if command.Name() != config.ActivityCommandName || !command.HasAlias(config.ActivityCommandAlias) {
		t.Fatalf("command = %q aliases %v", command.Name(), command.Aliases)
	}
	for name, count := range map[string]int{"list": 1, "get": 2} {
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
		for _, flag := range []string{"workspace", "project-id", "cursor", "per-page", "fields", "expand", "order-by"} {
			if child.Flags().Lookup(flag) == nil {
				t.Fatalf("%s missing --%s", name, flag)
			}
		}
	}
	for _, unsupported := range []string{"create", "update", "delete", "api-key", "bearer", "token"} {
		if command.Flags().Lookup(unsupported) != nil {
			t.Fatalf("activity exposed unsupported flag/command %q", unsupported)
		}
	}
	legacy, _, err := command.Find([]string{"legacy"})
	if err != nil || legacy == nil || !legacy.Hidden {
		t.Fatalf("legacy command = %#v, err=%v", legacy, err)
	}
	for name, count := range map[string]int{"list": 1, "get": 2} {
		child, _, err := legacy.Find([]string{name})
		if err != nil || child == nil {
			t.Fatalf("find legacy %s: %v", name, err)
		}
		if err := child.Args(child, make([]string, count)); err != nil {
			t.Fatalf("legacy %s rejected exact args: %v", name, err)
		}
	}
	if got := len(legacy.Commands()); got != 2 {
		t.Fatalf("legacy command count = %d, want 2", got)
	}
	for _, child := range legacy.Commands() {
		if child.Name() != "list" && child.Name() != "get" {
			t.Fatalf("legacy exposed unsupported operation %q", child.Name())
		}
	}
}

func TestActivityCommandsUseContextOverridesOutputAndRouteBoundaries(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeActivityClient{response: `{"results":[{"id":"activity-1"}]}`}
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "saved-team", Project: config.ProjectContext{ID: "saved-project"}}}
	var calls int
	clientFactory = func() (plane.Client, error) { calls++; return fake, nil }

	list := newListCommand(false)
	for name, value := range map[string]string{"cursor": "next", "per-page": "20", "fields": "id,verb", "expand": "actor", "order-by": "-created_at"} {
		if err := list.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	list.SetContext(context.Background())
	output := captureActivityStdout(t, func() {
		if err := runList(list, []string{"work/item"}); err != nil {
			t.Fatal(err)
		}
	})
	if calls != 1 || fake.routes[0] != "/workspaces/saved-team/projects/saved-project/work-items/work%2Fitem/activities/" {
		t.Fatalf("factory/routes = %d/%v", calls, fake.routes)
	}
	if fake.queries[0].Encode() != "cursor=next&expand=actor&fields=id%2Cverb&order_by=-created_at&per_page=20" {
		t.Fatalf("list query = %v", fake.queries[0])
	}
	if !strings.Contains(output, `"activity-1"`) {
		t.Fatalf("list output = %s", output)
	}

	get := newGetCommand(false)
	for name, value := range map[string]string{"workspace": "override-team", "project-id": "override-project", "cursor": "next", "per-page": "20", "fields": "id", "expand": "project", "order-by": "created_at"} {
		if err := get.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := runGet(get, []string{"work/item", "activity/item"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[1] != "/workspaces/override-team/projects/override-project/work-items/work%2Fitem/activities/activity%2Fitem/" {
		t.Fatalf("primary detail route = %q", fake.routes[1])
	}
	if fake.queries[1].Get("order_by") != "created_at" {
		t.Fatalf("detail query = %v", fake.queries[1])
	}
	if c.Context.Workspace != "saved-team" || c.Context.Project.ID != "saved-project" {
		t.Fatalf("override changed saved context: %#v", c.Context)
	}

	legacyList := newListCommand(true)
	if err := runLegacyList(legacyList, []string{"issue/item"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[2] != "/workspaces/saved-team/projects/saved-project/issues/issue%2Fitem/activities/" {
		t.Fatalf("legacy list route = %q", fake.routes[2])
	}
	legacyGet := newGetCommand(true)
	if err := runLegacyGet(legacyGet, []string{"issue/item", "activity/item"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[3] != "/workspaces/saved-team/projects/saved-project/issues/issue%2Fitem/activities/activity%2Fitem/" {
		t.Fatalf("legacy detail route = %q", fake.routes[3])
	}
}

func TestActivityCommandsRejectInvalidPageAndBlankOverridesBeforeFactory(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	var calls int
	c = &config.Config{Context: config.Context{Workspace: "team", Project: config.ProjectContext{ID: "project"}}}
	clientFactory = func() (plane.Client, error) { calls++; return &fakeActivityClient{}, nil }

	invalidPage := newListCommand(false)
	if err := invalidPage.Flags().Set("per-page", "101"); err != nil {
		t.Fatal(err)
	}
	if err := runList(invalidPage, []string{"work-item"}); err == nil || calls != 0 {
		t.Fatalf("invalid page result=%v factory calls=%d", err, calls)
	}
	blank := newListCommand(false)
	if err := blank.Flags().Set("project-id", ""); err != nil {
		t.Fatal(err)
	}
	if err := runList(blank, []string{"work-item"}); err == nil || calls != 0 {
		t.Fatalf("blank override result=%v factory calls=%d", err, calls)
	}
}

func TestActivityHelpIsSafeAndExplicit(t *testing.T) {
	helpText := strings.ToLower((&help.ActivityCmd{}).Long())
	for _, phrase := range []string{"activity", "activities", "legacy", "work-item", "issue", "context.workspace", "context.project.id", "cursor", "per-page", "fields", "expand", "order-by", "json", "yaml", "gron", "text", "table", "raw", "read-only"} {
		if !strings.Contains(helpText, phrase) {
			t.Fatalf("help omitted %q", phrase)
		}
	}
	for _, secret := range []string{"api-key", "bearer", "token", "presigned", "invitation"} {
		if strings.Contains(helpText, secret) {
			t.Fatalf("help contains prohibited %q", secret)
		}
	}
}

type fakeActivityClient struct {
	routes   []string
	methods  []string
	queries  []url.Values
	response string
	status   int
}

func (f *fakeActivityClient) Do(_ context.Context, method, route string, query url.Values, body any, _ http.Header, destination any) (plane.Response, error) {
	if body != nil {
		return plane.Response{}, nil
	}
	f.routes = append(f.routes, route)
	f.methods = append(f.methods, method)
	clone := make(url.Values, len(query))
	for key, values := range query {
		clone[key] = append([]string(nil), values...)
	}
	f.queries = append(f.queries, clone)
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

func captureActivityStdout(t *testing.T, function func()) string {
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

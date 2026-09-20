package link

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

	"github.com/spf13/cobra"
)

func TestLinkCommandRegistrationFlagsAndHiddenLegacy(t *testing.T) {
	command := Init(&config.Config{}, nil)
	if command.Name() != config.LinkCommandName || !command.HasAlias(config.LinkCommandAlias) {
		t.Fatalf("command = %q aliases %v", command.Name(), command.Aliases)
	}
	wantArgs := map[string]int{"list": 1, "create": 1, "get": 2, "update": 2, "delete": 2}
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
	get, _, _ := command.Find([]string{"get"})
	for _, flag := range []string{"cursor", "per-page", "fields", "expand"} {
		if get.Flags().Lookup(flag) == nil {
			t.Fatalf("get missing --%s", flag)
		}
	}
	if get.Flags().Lookup("order-by") != nil {
		t.Fatal("get exposed unsupported --order-by")
	}
	create, _, _ := command.Find([]string{"create"})
	if create.Flags().Lookup("url") == nil || create.Flags().Lookup("title") == nil {
		t.Fatal("create did not expose url/title")
	}
	if create.Flags().Lookup("url").Annotations[cobra.BashCompOneRequiredFlag] == nil {
		t.Fatal("create URL is not required")
	}
	update, _, _ := command.Find([]string{"update"})
	for _, flag := range []string{"url", "title"} {
		if update.Flags().Lookup(flag) == nil {
			t.Fatalf("update missing --%s", flag)
		}
	}
	if command.Flags().Lookup("api-key") != nil {
		t.Fatal("link command exposed credential flags")
	}
	legacy, _, err := command.Find([]string{"legacy"})
	if err != nil || legacy == nil || !legacy.Hidden {
		t.Fatalf("legacy command = %#v, err=%v", legacy, err)
	}
	for name, count := range wantArgs {
		child, _, err := legacy.Find([]string{name})
		if err != nil || child == nil {
			t.Fatalf("find legacy %s: %v", name, err)
		}
		if err := child.Args(child, make([]string, count)); err != nil {
			t.Fatalf("legacy %s rejected exact args: %v", name, err)
		}
	}
}

func TestLinkCommandsUseContextOverridesAndCentralOutput(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeLinkClient{response: `{"id":"link-1","url":"https://example.test","results":[]}`}
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "saved-team", Project: config.ProjectContext{ID: "saved-project"}}}
	var calls int
	clientFactory = func() (plane.Client, error) { calls++; return fake, nil }

	list := newListCommand(false)
	for name, value := range map[string]string{"cursor": "next", "per-page": "20", "fields": "id,url", "expand": "project", "order-by": "-created_at"} {
		if err := list.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	list.SetContext(context.Background())
	output := captureLinkStdout(t, func() {
		if err := runList(list, []string{"work-item"}); err != nil {
			t.Fatal(err)
		}
	})
	if calls != 1 || fake.routes[0] != "/workspaces/saved-team/projects/saved-project/work-items/work-item/links/" {
		t.Fatalf("factory/routes = %d/%v", calls, fake.routes)
	}
	if fake.queries[0].Encode() != "cursor=next&expand=project&fields=id%2Curl&order_by=-created_at&per_page=20" {
		t.Fatalf("list query = %v", fake.queries[0])
	}
	if !strings.Contains(output, `"results":[]`) {
		t.Fatalf("list output = %s", output)
	}

	fake.status = http.StatusCreated
	override := newCreateCommand(false)
	for name, value := range map[string]string{"workspace": "override-team", "project-id": "override-project", "url": "https://example.test", "title": ""} {
		if err := override.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := runCreate(override, []string{"work-item"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[1] != "/workspaces/override-team/projects/override-project/work-items/work-item/links/" {
		t.Fatalf("override route = %q", fake.routes[1])
	}
	if string(fake.body) != `{"url":"https://example.test","title":""}` {
		t.Fatalf("create body = %s", fake.body)
	}
	if c.Context.Workspace != "saved-team" || c.Context.Project.ID != "saved-project" {
		t.Fatalf("override changed saved context: %#v", c.Context)
	}
}

func TestLinkUpdateGetLegacyAndDeleteContracts(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	fake := &fakeLinkClient{response: `{"id":"link-1","title":"updated","results":[{"id":"link-1"}]}`}
	c = &config.Config{OutputFormat: "raw", Context: config.Context{Workspace: "team", Project: config.ProjectContext{ID: "project"}}}
	clientFactory = func() (plane.Client, error) { return fake, nil }

	update := newUpdateCommand(false)
	if err := update.Flags().Set("title", "updated"); err != nil {
		t.Fatal(err)
	}
	if err := runUpdate(update, []string{"work-item", "link"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[0] != "/workspaces/team/projects/project/work-items/work-item/links/link/" || string(fake.body) != `{"title":"updated"}` {
		t.Fatalf("update route/body = %q / %s", fake.routes[0], fake.body)
	}

	get := newGetCommand(false)
	for name, value := range map[string]string{"cursor": "next", "per-page": "20", "fields": "id", "expand": "project"} {
		if err := get.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := runGet(get, []string{"work-item", "link"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[1] != "/workspaces/team/projects/project/work-items/work-item/links/link/" || fake.queries[1].Get("order_by") != "" {
		t.Fatalf("get route/query = %q / %v", fake.routes[1], fake.queries[1])
	}

	legacy := newListCommand(true)
	if err := runLegacyList(legacy, []string{"issue"}); err != nil {
		t.Fatal(err)
	}
	if fake.routes[2] != "/workspaces/team/projects/project/issues/issue/links/" {
		t.Fatalf("legacy route = %q", fake.routes[2])
	}

	deleteCommand := newDeleteCommand(false)
	fake.status = http.StatusNoContent
	output := captureLinkStdout(t, func() {
		if err := runDelete(deleteCommand, []string{"work-item", "link"}); err != nil {
			t.Fatal(err)
		}
	})
	if output != "" || fake.routes[3] != "/workspaces/team/projects/project/work-items/work-item/links/link/" {
		t.Fatalf("delete output/route = %q / %q", output, fake.routes[3])
	}
}

func TestLinkCommandRejectsInvalidPageAndBlankOverridesBeforeFactory(t *testing.T) {
	previousConfig, previousFactory := c, clientFactory
	t.Cleanup(func() { c, clientFactory = previousConfig, previousFactory })
	var calls int
	c = &config.Config{Context: config.Context{Workspace: "team", Project: config.ProjectContext{ID: "project"}}}
	clientFactory = func() (plane.Client, error) { calls++; return &fakeLinkClient{}, nil }

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

func TestLinkHelpIsSafeAndExplicit(t *testing.T) {
	helpText := strings.ToLower((&help.LinkCmd{}).Long())
	for _, phrase := range []string{"link", "links", "legacy", "work-item", "issue", "context.workspace", "context.project.id", "cursor", "per-page", "fields", "expand", "order-by", "url", "title", "json", "yaml", "gron", "text", "table", "raw"} {
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

type fakeLinkClient struct {
	routes   []string
	queries  []url.Values
	body     []byte
	response string
	status   int
}

func (f *fakeLinkClient) Do(_ context.Context, _ string, route string, query url.Values, body any, _ http.Header, destination any) (plane.Response, error) {
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

func captureLinkStdout(t *testing.T, function func()) string {
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

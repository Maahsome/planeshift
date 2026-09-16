package projectlabel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"planeshift/config"
	"planeshift/plane"
)

func TestProjectLabelCommandsHaveExactArgumentsAndDocumentedFlags(t *testing.T) {
	command := Init(&config.Config{}, nil)
	if command.Name() != config.ProjectLabelCommandName || !command.HasAlias(config.ProjectLabelCommandAlias) {
		t.Fatalf("command = %q aliases=%v", command.Name(), command.Aliases)
	}
	if command.Flags().NFlag() != 0 {
		t.Fatalf("parent has operation flags: %v", command.Flags().FlagUsages())
	}
	wantCommands := map[string]int{"create": 1, "list": 1, "get": 2, "update": 2, "delete": 2}
	for name, argumentCount := range wantCommands {
		child, _, err := command.Find([]string{name})
		if err != nil || child == nil {
			t.Fatalf("find %s: command=%v err=%v", name, child, err)
		}
		if err := child.Args(child, make([]string, argumentCount-1)); err == nil {
			t.Fatalf("%s accepted %d positional arguments", name, argumentCount-1)
		}
		if err := child.Args(child, make([]string, argumentCount)); err != nil {
			t.Fatalf("%s rejected exact arguments: %v", name, err)
		}
	}

	list, _, _ := command.Find([]string{"list"})
	for _, name := range []string{"cursor", "per-page", "fields", "expand", "order-by"} {
		if list.Flags().Lookup(name) == nil {
			t.Fatalf("list missing --%s", name)
		}
	}
	for _, name := range []string{"name", "description", "color", "sort-order"} {
		if list.Flags().Lookup(name) != nil {
			t.Fatalf("list has unsupported --%s", name)
		}
	}
	for _, operation := range []string{"get", "delete"} {
		child, _, _ := command.Find([]string{operation})
		for _, name := range []string{"cursor", "per-page", "fields", "expand", "order-by", "name", "description", "color", "sort-order"} {
			if child.Flags().Lookup(name) != nil {
				t.Fatalf("%s has unsupported --%s", operation, name)
			}
		}
	}
	for _, operation := range []string{"create", "update"} {
		child, _, _ := command.Find([]string{operation})
		for _, name := range []string{"name", "description", "color", "sort-order"} {
			if child.Flags().Lookup(name) == nil {
				t.Fatalf("%s missing --%s", operation, name)
			}
		}
	}
}

func TestProjectLabelHelpIsSafeAndComplete(t *testing.T) {
	command := Init(&config.Config{}, nil)
	for _, phrase := range []string{
		"workspace_slug", "label_id", "project-labels", "create", "list", "get", "update", "delete",
		"cursor", "per-page", "fields", "expand", "order-by", "--name", "--description", "--color", "--sort-order",
		"204", "json", "yaml", "gron", "text", "table", "raw",
	} {
		if !strings.Contains(command.Long, phrase) {
			t.Fatalf("help omitted %q: %s", phrase, command.Long)
		}
	}
	for _, prohibited := range []string{"api-key", "bearer", "invitation", "presigned", "upload", "token"} {
		if strings.Contains(strings.ToLower(command.Long), prohibited) {
			t.Fatalf("help contains prohibited %q", prohibited)
		}
	}
}

func TestProjectLabelFactoryIsLazyAndListValidationPrecedesFactory(t *testing.T) {
	previousConfig, previousFactory := commandConfig, clientFactory
	t.Cleanup(func() { commandConfig, clientFactory = previousConfig, previousFactory })
	var calls int
	fake := &fakeProjectLabelClient{responseJSON: `{"results":[]}`}
	Init(&config.Config{OutputFormat: "raw"}, func() (plane.Client, error) {
		calls++
		return fake, nil
	})
	if calls != 0 {
		t.Fatal("factory invoked while constructing hierarchy")
	}
	invalid := newListCommand()
	if err := invalid.Flags().Set("per-page", "0"); err != nil {
		t.Fatal(err)
	}
	if err := runList(invalid, []string{"team"}); err == nil || calls != 0 {
		t.Fatalf("invalid list = %v, factory calls=%d", err, calls)
	}
}

func TestProjectLabelCreateAndUpdatePreserveExplicitEmptyAndZeroValues(t *testing.T) {
	previousConfig, previousFactory := commandConfig, clientFactory
	t.Cleanup(func() { commandConfig, clientFactory = previousConfig, previousFactory })
	fake := &fakeProjectLabelClient{responseJSON: `{"id":"label-1","name":"Label"}`}
	Init(&config.Config{OutputFormat: "raw"}, plane.StaticClientFactory(fake))

	create := newCreateCommand()
	for name, value := range map[string]string{
		"name": "Label", "description": "", "color": "", "sort-order": "0",
	} {
		if err := create.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := runCreate(create, []string{"team"}); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(fake.lastBody)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"name":"Label","description":"","color":"","sort_order":0}` {
		t.Fatalf("create body = %s", body)
	}

	fake.lastBody = nil
	update := newUpdateCommand()
	for name, value := range map[string]string{"description": "", "sort-order": "0"} {
		if err := update.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := runUpdate(update, []string{"team", "label-1"}); err != nil {
		t.Fatal(err)
	}
	body, err = json.Marshal(fake.lastBody)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"description":"","sort_order":0}` {
		t.Fatalf("update body = %s", body)
	}
}

func TestProjectLabelListBuildsAllQueriesAndOutputUsesConfig(t *testing.T) {
	previousConfig, previousFactory := commandConfig, clientFactory
	t.Cleanup(func() { commandConfig, clientFactory = previousConfig, previousFactory })
	fake := &fakeProjectLabelClient{responseJSON: `{"next_cursor":"next","total_count":1,"extra_stats":null,"results":[{"id":"label-1","name":"Label"}],"future":true}`}
	Init(&config.Config{OutputFormat: "raw"}, plane.StaticClientFactory(fake))
	list := newListCommand()
	for name, value := range map[string]string{
		"cursor": "next:1", "per-page": "20", "fields": "id,name", "expand": "projects", "order-by": "-sort_order",
	} {
		if err := list.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	output := captureStdout(t, func() {
		if err := runList(list, []string{"team"}); err != nil {
			t.Fatal(err)
		}
	})
	wantQuery := url.Values{
		"cursor": {"next:1"}, "per_page": {"20"}, "fields": {"id,name"}, "expand": {"projects"}, "order_by": {"-sort_order"},
	}
	if !reflect.DeepEqual(fake.query, wantQuery) {
		t.Fatalf("list query = %v, want %v", fake.query, wantQuery)
	}
	if !strings.Contains(output, `"future":true`) {
		t.Fatalf("configured raw output omitted unknown page data: %q", output)
	}
}

func TestProjectLabelEmptyUpdateSendsEmptyObjectAndDeleteIsQuiet(t *testing.T) {
	previousConfig, previousFactory := commandConfig, clientFactory
	t.Cleanup(func() { commandConfig, clientFactory = previousConfig, previousFactory })
	fake := &fakeProjectLabelClient{statusCode: http.StatusOK, responseJSON: `{"id":"label-1"}`}
	Init(&config.Config{OutputFormat: "json"}, plane.StaticClientFactory(fake))
	update := newUpdateCommand()
	if err := runUpdate(update, []string{"team", "label-1"}); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(fake.lastBody)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{}` {
		t.Fatalf("empty update body = %s", body)
	}
	fake.lastBody = nil
	fake.statusCode = http.StatusNoContent
	output := captureStdout(t, func() {
		if err := runDelete(newDeleteCommand(), []string{"team", "label-1"}); err != nil {
			t.Fatal(err)
		}
	})
	if output != "" || fake.lastBody != nil || fake.lastDestination != nil {
		t.Fatalf("delete output/body/destination = %q/%#v/%#v", output, fake.lastBody, fake.lastDestination)
	}
}

type fakeProjectLabelClient struct {
	calls           int
	method          string
	route           string
	query           url.Values
	lastBody        any
	lastDestination any
	responseJSON    string
	statusCode      int
}

func (f *fakeProjectLabelClient) Do(_ context.Context, method, route string, query url.Values, body any, _ http.Header, destination any) (plane.Response, error) {
	f.calls++
	f.method, f.route, f.query = method, route, query
	f.lastBody, f.lastDestination = body, destination
	status := f.statusCode
	if status == 0 {
		status = http.StatusOK
	}
	if destination != nil && f.responseJSON != "" {
		if err := json.Unmarshal([]byte(f.responseJSON), destination); err != nil {
			return plane.Response{StatusCode: status}, err
		}
	}
	return plane.Response{StatusCode: status}, nil
}

func captureStdout(t *testing.T, function func()) string {
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
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return string(data)
}

var _ plane.Client = (*fakeProjectLabelClient)(nil)

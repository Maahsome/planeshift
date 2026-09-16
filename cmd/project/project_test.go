package project

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

	"github.com/spf13/cobra"
)

func TestProjectCommandsAreRegisteredWithExactArgumentsAndFlags(t *testing.T) {
	command := Init(&config.Config{}, nil)
	if command.Name() != config.ProjectCommandName {
		t.Fatalf("project command name = %q, want %q", command.Name(), config.ProjectCommandName)
	}
	if !command.HasAlias(config.ProjectCommandAlias) {
		t.Fatalf("project command aliases = %v, want %q", command.Aliases, config.ProjectCommandAlias)
	}
	if command.Flags().NFlag() != 0 {
		t.Fatalf("project parent has operation-specific flags: %v", command.Flags().FlagUsages())
	}
	wantCommands := map[string]int{
		"list": 1, "create": 1, "create-template": 1, "get": 2,
		"update": 2, "archive": 2, "unarchive": 2, "delete": 2,
	}
	for name, argumentCount := range wantCommands {
		child, _, err := command.Find([]string{name})
		if err != nil || child == nil {
			t.Fatalf("find %s: command=%v err=%v", name, child, err)
		}
		if err := child.Args(child, make([]string, argumentCount-1)); err == nil {
			t.Fatalf("%s accepted %d positional argument(s), want exact %d", name, argumentCount-1, argumentCount)
		}
		if err := child.Args(child, make([]string, argumentCount)); err != nil {
			t.Fatalf("%s rejected exact positional arguments: %v", name, err)
		}
	}
	if features, _, err := command.Find([]string{"features"}); err == nil && features != nil && features != command {
		t.Fatalf("unsupported features command is still registered: %v", features)
	}

	list, _, _ := command.Find([]string{"list"})
	for _, name := range []string{"cursor", "per-page", "fields", "expand", "order-by"} {
		if list.Flags().Lookup(name) == nil {
			t.Fatalf("list missing --%s", name)
		}
	}
	for _, name := range []string{"cursor", "per-page", "fields", "expand", "order-by"} {
		for _, operation := range []string{"get", "archive", "unarchive", "delete"} {
			child, _, _ := command.Find([]string{operation})
			if child.Flags().Lookup(name) != nil {
				t.Fatalf("%s inherited list flag --%s", operation, name)
			}
		}
	}
	create, _, _ := command.Find([]string{"create"})
	for _, name := range []string{"name", "identifier", "description", "icon-prop", "intake-view", "guest-view-all-features", "external-source", "is-time-tracking-enabled"} {
		if create.Flags().Lookup(name) == nil {
			t.Fatalf("create missing --%s", name)
		}
	}
	template, _, _ := command.Find([]string{"create-template"})
	for _, name := range []string{"template-id", "name", "identifier", "description", "network", "project-lead"} {
		if template.Flags().Lookup(name) == nil {
			t.Fatalf("create-template missing --%s", name)
		}
	}
}

func TestProjectHelpDocumentsSafeLifecycleAndDynamicOutput(t *testing.T) {
	command := Init(&config.Config{}, nil)
	help := command.Long
	for _, phrase := range []string{"workspace slug", "project ID", "cursor pagination", "archive", "unarchive", "204", "icon-prop", "planeshift project", "planeshift projects"} {
		if !strings.Contains(help, phrase) {
			t.Fatalf("project help omitted %q: %s", phrase, help)
		}
	}
	for _, secret := range []string{"api-key", "bearer", "invitation", "presigned", "upload"} {
		if strings.Contains(strings.ToLower(help), secret) {
			t.Fatalf("project help contains prohibited %q", secret)
		}
	}
}

func TestProjectCommandsResolveFactoryOnlyDuringExecution(t *testing.T) {
	previousConfig := c
	previousFactory := clientFactory
	t.Cleanup(func() {
		c = previousConfig
		clientFactory = previousFactory
	})

	var calls int
	fake := &fakeProjectClient{responseJSON: `{"id":"project-1","name":"Project","identifier":"PRJ"}`}
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) {
		calls++
		return fake, nil
	}
	if calls != 0 {
		t.Fatal("factory was invoked during setup")
	}

	command := &cobra.Command{}
	command.SetContext(context.Background())
	output := captureStdout(t, func() {
		if err := runGet(command, []string{"team", "project-1"}); err != nil {
			t.Fatal(err)
		}
	})
	if calls != 1 || fake.calls != 1 {
		t.Fatalf("factory/client calls = %d/%d, want 1/1", calls, fake.calls)
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("raw project output = %q: %v", output, err)
	}
	if raw["id"] != "project-1" || raw["name"] != "Project" || raw["identifier"] != "PRJ" {
		t.Fatalf("raw project output = %q", output)
	}
}

func TestProjectBodyCommandsUseConfiguredProjectOutputFormats(t *testing.T) {
	previousConfig := c
	previousFactory := clientFactory
	t.Cleanup(func() {
		c = previousConfig
		clientFactory = previousFactory
	})

	const singleJSON = `{"id":"project-1","identifier":"PRJ","name":"Project X","workspace":"workspace-1","network":2,"description_text":null,"icon_prop":{"name":"rocket"},"default_state":null,"future_field":{"keep":true}}`
	fake := &fakeProjectClient{responseJSON: singleJSON}
	clientFactory = func() (plane.Client, error) { return fake, nil }

	t.Run("get defaults to JSON", func(t *testing.T) {
		c = &config.Config{}
		command := &cobra.Command{}
		command.SetContext(context.Background())
		output := captureStdout(t, func() {
			if err := runGet(command, []string{"team", "project-1"}); err != nil {
				t.Fatal(err)
			}
		})
		if c.OutputFormat != "json" {
			t.Fatalf("default output format = %q, want json", c.OutputFormat)
		}
		assertProjectOutputFields(t, output)
	})

	t.Run("get preserves raw JSON", func(t *testing.T) {
		c = &config.Config{OutputFormat: "raw"}
		command := &cobra.Command{}
		command.SetContext(context.Background())
		output := captureStdout(t, func() {
			if err := runGet(command, []string{"team", "project-1"}); err != nil {
				t.Fatal(err)
			}
		})
		assertProjectOutputFields(t, output)
		if !strings.Contains(output, `"description_text":null`) || !strings.Contains(output, `"future_field":{"keep":true}`) {
			t.Fatalf("raw output lost nullable/unknown fields: %q", output)
		}
	})

	t.Run("create renders text summary", func(t *testing.T) {
		c = &config.Config{OutputFormat: "text"}
		command := &cobra.Command{}
		addCreateFlags(command)
		if err := command.Flags().Set("name", "Project X"); err != nil {
			t.Fatal(err)
		}
		if err := command.Flags().Set("identifier", "PRJ"); err != nil {
			t.Fatal(err)
		}
		output := captureStdout(t, func() {
			if err := runCreate(command, []string{"team"}); err != nil {
				t.Fatal(err)
			}
		})
		for _, value := range []string{"ID", "IDENTIFIER", "NAME", "WORKSPACE", "NETWORK", "project-1", "PRJ", "Project X", "workspace-1", "2"} {
			if !strings.Contains(output, value) {
				t.Fatalf("text output omitted %q: %q", value, output)
			}
		}
		if strings.Contains(output, "future_field") || strings.Contains(output, "rocket") {
			t.Fatalf("text output dumped dynamic fields: %q", output)
		}
	})

	t.Run("list renders ordered table rows", func(t *testing.T) {
		c = &config.Config{OutputFormat: "table"}
		fake.responseJSON = `{"next_cursor":"next","results":[{"id":"project-1","identifier":"ONE","name":"First","workspace":null,"network":null},{"id":"project-2","identifier":"TWO","name":"Second","workspace":"workspace-2","network":0}]}`
		command := &cobra.Command{}
		addListFlags(command)
		command.SetContext(context.Background())
		output := captureStdout(t, func() {
			if err := runList(command, []string{"team"}); err != nil {
				t.Fatal(err)
			}
		})
		first := strings.Index(output, "First")
		second := strings.Index(output, "Second")
		if first == -1 || second == -1 || first >= second {
			t.Fatalf("list output order = %q", output)
		}
		for _, value := range []string{"ID", "IDENTIFIER", "NAME", "WORKSPACE", "NETWORK", "project-1", "ONE", "project-2", "TWO", "workspace-2", "0"} {
			if !strings.Contains(output, value) {
				t.Fatalf("table output omitted %q: %q", value, output)
			}
		}
		if strings.Contains(output, "next_cursor") {
			t.Fatalf("table output leaked page metadata: %q", output)
		}
	})
}

func assertProjectOutputFields(t *testing.T, output string) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &fields); err != nil {
		t.Fatalf("project output is not JSON: %q: %v", output, err)
	}
	for _, key := range []string{"id", "identifier", "name", "description_text", "icon_prop", "default_state", "future_field"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("project output omitted %q: %s", key, output)
		}
	}
}

func TestProjectListValidationHappensBeforeFactoryAndOutputUsesConfig(t *testing.T) {
	previousConfig := c
	previousFactory := clientFactory
	t.Cleanup(func() {
		c = previousConfig
		clientFactory = previousFactory
	})

	var calls int
	c = &config.Config{OutputFormat: "json"}
	clientFactory = func() (plane.Client, error) {
		calls++
		return &fakeProjectClient{}, nil
	}
	invalid := &cobra.Command{}
	invalid.Flags().Int("per-page", 0, "")
	if err := invalid.Flags().Set("per-page", "0"); err != nil {
		t.Fatal(err)
	}
	if err := runList(invalid, []string{"team"}); err == nil || calls != 0 {
		t.Fatalf("invalid list result = %v, factory calls = %d", err, calls)
	}

	fake := &fakeProjectClient{responseJSON: `{"results":[],"next_cursor":null,"future":true}`}
	clientFactory = func() (plane.Client, error) { return fake, nil }
	valid := &cobra.Command{}
	valid.SetContext(context.Background())
	valid.Flags().String("cursor", "cursor:1", "")
	valid.Flags().Int("per-page", 20, "")
	valid.Flags().String("fields", "id,name", "")
	valid.Flags().String("expand", "members", "")
	valid.Flags().String("order-by", "-created_at", "")
	output := captureStdout(t, func() {
		if err := runList(valid, []string{"team"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, `"future"`) {
		t.Fatalf("list output did not use RawJSON/config dispatch: %q", output)
	}
	if fake.query.Get("order_by") != "-created_at" || fake.query.Get("per_page") != "20" {
		t.Fatalf("list query = %v", fake.query)
	}
}

func TestProjectLifecycle204CommandsAreQuiet(t *testing.T) {
	previousConfig := c
	previousFactory := clientFactory
	t.Cleanup(func() {
		c = previousConfig
		clientFactory = previousFactory
	})
	fake := &fakeProjectClient{statusCode: http.StatusNoContent}
	c = &config.Config{OutputFormat: "json"}
	clientFactory = func() (plane.Client, error) { return fake, nil }
	command := &cobra.Command{}
	command.SetContext(context.Background())
	for name, run := range map[string]func(*cobra.Command, []string) error{
		"archive": runArchive, "unarchive": runUnarchive, "delete": runDelete,
	} {
		output := captureStdout(t, func() {
			if err := run(command, []string{"team", "project"}); err != nil {
				t.Fatalf("%s: %v", name, err)
			}
		})
		if output != "" {
			t.Fatalf("%s wrote output %q for 204", name, output)
		}
	}
	if fake.calls != 3 || fake.lastBody != nil || fake.lastDestination != nil {
		t.Fatalf("lifecycle calls/body/destination = %d/%#v/%#v", fake.calls, fake.lastBody, fake.lastDestination)
	}
}

func TestProjectCreateAndUpdatePreserveChangedFalseZeroAndJSONFlags(t *testing.T) {
	previousConfig := c
	previousFactory := clientFactory
	t.Cleanup(func() {
		c = previousConfig
		clientFactory = previousFactory
	})
	fake := &fakeProjectClient{responseJSON: `{"id":"project"}`}
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) { return fake, nil }

	create := &cobra.Command{}
	addCreateFlags(create)
	for name, value := range map[string]string{
		"name": "Project", "identifier": "PRJ", "module-view": "false", "archive-in": "0",
		"icon-prop": `{"color":"blue"}`,
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
	var createBody map[string]any
	if err := json.Unmarshal(body, &createBody); err != nil {
		t.Fatal(err)
	}
	if createBody["module_view"] != false || createBody["archive_in"] != float64(0) || createBody["icon_prop"].(map[string]any)["color"] != "blue" {
		t.Fatalf("create body omitted changed values: %s", body)
	}

	fake.lastBody = nil
	update := &cobra.Command{}
	addUpdateFlags(update)
	for name, value := range map[string]string{
		"cycle-view": "false", "close-in": "0", "default-state": "", "icon-prop": "null",
	} {
		if err := update.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := runUpdate(update, []string{"team", "project"}); err != nil {
		t.Fatal(err)
	}
	body, err = json.Marshal(fake.lastBody)
	if err != nil {
		t.Fatal(err)
	}
	var updateBody map[string]any
	if err := json.Unmarshal(body, &updateBody); err != nil {
		t.Fatal(err)
	}
	if updateBody["cycle_view"] != false || updateBody["close_in"] != float64(0) || updateBody["default_state"] != "" || updateBody["icon_prop"] != nil {
		t.Fatalf("update body omitted changed values: %s", body)
	}
}

func TestProjectInvalidIconIsRejectedBeforeFactory(t *testing.T) {
	previousConfig := c
	previousFactory := clientFactory
	t.Cleanup(func() {
		c = previousConfig
		clientFactory = previousFactory
	})
	var calls int
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) {
		calls++
		return &fakeProjectClient{}, nil
	}
	command := &cobra.Command{}
	addCreateFlags(command)
	_ = command.Flags().Set("name", "Project")
	_ = command.Flags().Set("identifier", "PRJ")
	_ = command.Flags().Set("icon-prop", "not-json")
	if err := runCreate(command, []string{"team"}); err == nil || calls != 0 {
		t.Fatalf("invalid icon result = %v, factory calls = %d", err, calls)
	}
}

type fakeProjectClient struct {
	calls           int
	query           url.Values
	responseJSON    string
	statusCode      int
	lastBody        any
	lastDestination any
}

func (f *fakeProjectClient) Do(_ context.Context, method, route string, query url.Values, body any, _ http.Header, destination any) (plane.Response, error) {
	f.calls++
	f.query = query
	f.lastBody = body
	f.lastDestination = destination
	if method == "" || route == "" {
		return plane.Response{}, nil
	}
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

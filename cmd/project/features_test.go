package project

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"planeshift/config"
	"planeshift/plane"
	projectfeatures "planeshift/projectfeatures"
)

func TestProjectFeaturesCommandsAreRegisteredWithExactArgumentsAndFlags(t *testing.T) {
	command := Init(&config.Config{}, nil)
	features, _, err := command.Find([]string{"features"})
	if err != nil || features == nil {
		t.Fatalf("features command = %v, err=%v", features, err)
	}
	if features.Parent() != command || features.Flags().NFlag() != 0 {
		t.Fatalf("features parent = %v, flags=%v", features.Parent(), features.Flags().FlagUsages())
	}
	if err := features.Args(features, nil); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"get", "update"} {
		child, _, err := command.Find([]string{"features", name})
		if err != nil || child == nil {
			t.Fatalf("find features %s: command=%v err=%v", name, child, err)
		}
		if err := child.Args(child, []string{"workspace"}); err == nil {
			t.Fatalf("features %s accepted one positional argument", name)
		}
		if err := child.Args(child, []string{"workspace", "project"}); err != nil {
			t.Fatalf("features %s rejected exact arguments: %v", name, err)
		}
	}

	update, _, _ := command.Find([]string{"features", "update"})
	for _, name := range []string{"epics", "modules", "cycles", "views", "pages", "intakes", "work-item-types"} {
		if update.Flags().Lookup(name) == nil {
			t.Fatalf("update missing --%s", name)
		}
	}
	for _, name := range []string{"cursor", "per-page", "fields", "expand", "order-by"} {
		if update.Flags().Lookup(name) != nil {
			t.Fatalf("update unexpectedly exposes --%s", name)
		}
	}
	get, _, _ := command.Find([]string{"features", "get"})
	if get.Flags().NFlag() != 0 {
		t.Fatalf("get has unsupported flags: %s", get.Flags().FlagUsages())
	}
}

func TestProjectFeaturesHelpIsCompleteAndSafe(t *testing.T) {
	command := Init(&config.Config{}, nil)
	features, _, err := command.Find([]string{"features"})
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{
		"get workspace_slug project_id", "update workspace_slug project_id", "partial",
		"--epics", "--modules", "--cycles", "--views", "--pages", "--intakes",
		"--work-item-types", "--flag=false", "-o/--output", "json", "yaml", "gron", "text", "table", "raw",
	} {
		if !strings.Contains(features.Long, phrase) {
			t.Fatalf("feature help omitted %q: %s", phrase, features.Long)
		}
	}
	for _, secret := range []string{"api-key", "bearer", "invitation", "presigned", "upload", "token"} {
		if strings.Contains(strings.ToLower(features.Long), secret) {
			t.Fatalf("feature help contains prohibited %q", secret)
		}
	}
}

func TestProjectFeaturesCommandsUseLazyFactoryAndConfiguredOutput(t *testing.T) {
	previousConfig := c
	previousFactory := clientFactory
	t.Cleanup(func() {
		c = previousConfig
		clientFactory = previousFactory
	})

	fake := &featureClientDouble{responseJSON: `{"epics":true,"modules":false,"cycles":true,"views":false,"pages":true,"intakes":false,"work_item_types":true,"future":{"kept":true}}`}
	calls := 0
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) {
		calls++
		return fake, nil
	}
	command := newFeaturesGetCommand()
	if calls != 0 {
		t.Fatal("factory invoked while building get command")
	}
	command.SetContext(context.Background())
	output := captureStdout(t, func() {
		if err := runFeaturesGet(command, []string{"team", "project"}); err != nil {
			t.Fatal(err)
		}
	})
	if calls != 1 || fake.calls != 1 {
		t.Fatalf("factory/client calls = %d/%d, want 1/1", calls, fake.calls)
	}
	if fake.method != http.MethodGet || fake.route != "/workspaces/team/projects/project/features/" || fake.query != nil || fake.body != nil || fake.destination == nil {
		t.Fatalf("GET request = %#v", fake)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &fields); err != nil {
		t.Fatalf("feature output is not JSON: %q: %v", output, err)
	}
	if string(fields["future"]) != `{"kept":true}` || string(fields["epics"]) != "true" {
		t.Fatalf("feature output lost fields: %s", output)
	}
}

func TestProjectFeaturesUpdatePreservesChangedFalseAndOmitsUntouchedFlags(t *testing.T) {
	previousConfig := c
	previousFactory := clientFactory
	t.Cleanup(func() {
		c = previousConfig
		clientFactory = previousFactory
	})

	fake := &featureClientDouble{responseJSON: `{"epics":false,"modules":true}`}
	c = &config.Config{OutputFormat: "raw"}
	clientFactory = func() (plane.Client, error) { return fake, nil }
	command := newFeaturesUpdateCommand()
	if err := command.Flags().Set("epics", "false"); err != nil {
		t.Fatal(err)
	}
	if err := command.Flags().Set("modules", "true"); err != nil {
		t.Fatal(err)
	}
	command.SetContext(context.Background())
	if err := runFeaturesUpdate(command, []string{"team", "project"}); err != nil {
		t.Fatal(err)
	}
	if fake.method != http.MethodPatch || fake.route != "/workspaces/team/projects/project/features/" || fake.query != nil || fake.destination == nil {
		t.Fatalf("PATCH request = %#v", fake)
	}
	request, ok := fake.body.(projectfeatures.UpdateProjectFeaturesRequest)
	if !ok || request.Epics == nil || *request.Epics || request.Modules == nil || !*request.Modules {
		t.Fatalf("PATCH typed request = %#v", fake.body)
	}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cycles", "views", "pages", "intakes", "work_item_types"} {
		if strings.Contains(string(data), `"`+name+`"`) {
			t.Fatalf("untouched flag %q was serialized: %s", name, data)
		}
	}
}

func TestProjectFeaturesCommandsAreQuietForShared204(t *testing.T) {
	previousConfig := c
	previousFactory := clientFactory
	t.Cleanup(func() {
		c = previousConfig
		clientFactory = previousFactory
	})
	fake := &featureClientDouble{statusCode: http.StatusNoContent}
	c = &config.Config{OutputFormat: "json"}
	clientFactory = func() (plane.Client, error) { return fake, nil }

	get := newFeaturesGetCommand()
	get.SetContext(context.Background())
	getOutput := captureStdout(t, func() {
		if err := runFeaturesGet(get, []string{"team", "project"}); err != nil {
			t.Fatal(err)
		}
	})
	update := newFeaturesUpdateCommand()
	update.SetContext(context.Background())
	updateOutput := captureStdout(t, func() {
		if err := runFeaturesUpdate(update, []string{"team", "project"}); err != nil {
			t.Fatal(err)
		}
	})
	if getOutput != "" || updateOutput != "" {
		t.Fatalf("204 output = get %q, update %q", getOutput, updateOutput)
	}
}

type featureClientDouble struct {
	calls        int
	method       string
	route        string
	query        url.Values
	body         any
	destination  any
	statusCode   int
	responseJSON string
}

func (f *featureClientDouble) Do(_ context.Context, method, route string, query url.Values, body any, _ http.Header, destination any) (plane.Response, error) {
	f.calls++
	f.method = method
	f.route = route
	f.query = query
	f.body = body
	f.destination = destination
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

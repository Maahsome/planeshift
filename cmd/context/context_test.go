package context

import (
	stdcontext "context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"planeshift/config"
	"planeshift/plane"
	projectresource "planeshift/projects"
)

type fakePrompt struct {
	events       *[]string
	workspace    string
	project      string
	workspaceErr error
	projectErr   error
	options      []string
}

func (p *fakePrompt) Workspace() (string, error) {
	if p.events != nil {
		*p.events = append(*p.events, "workspace-prompt")
	}
	return p.workspace, p.workspaceErr
}

func (p *fakePrompt) Project(options []string) (string, error) {
	if p.events != nil {
		*p.events = append(*p.events, "project-prompt")
	}
	p.options = append([]string(nil), options...)
	return p.project, p.projectErr
}

type fakePlaneClient struct {
	events *[]string
	page   projectresource.ProjectPage
	err    error
}

func (c *fakePlaneClient) Do(_ stdcontext.Context, _ string, _ string, _ url.Values, _ any, _ http.Header, destination any) (plane.Response, error) {
	if c.events != nil {
		*c.events = append(*c.events, "project-list")
	}
	if c.err != nil {
		return plane.Response{}, c.err
	}
	page, ok := destination.(*projectresource.ProjectPage)
	if !ok {
		return plane.Response{}, errors.New("unexpected destination")
	}
	*page = c.page
	return plane.Response{}, nil
}

func projectPage(values ...projectresource.Project) projectresource.ProjectPage {
	return projectresource.ProjectPage{CursorPage: plane.CursorPage[projectresource.Project]{Results: values}}
}

func TestInitRegistersContextTreeWithoutConstructingPlaneClient(t *testing.T) {
	conf := &config.Config{}
	factoryCalls := 0
	command := Init(conf, func() (plane.Client, error) {
		factoryCalls++
		return nil, nil
	}, nil, &fakePrompt{})

	if command.Name() != config.ContextCommandName || command.Parent() != nil {
		t.Fatalf("context command = %v", command)
	}
	if command.Args(command, []string{"unexpected"}) == nil {
		t.Fatal("context accepted positional arguments")
	}
	for _, name := range []string{"set", "get", "prompt"} {
		child, _, err := command.Find([]string{name})
		if err != nil || child == nil || child.Parent() != command {
			t.Fatalf("%s registration = command=%v err=%v", name, child, err)
		}
		if child.Args(child, []string{"unexpected"}) == nil {
			t.Fatalf("context %s accepted positional arguments", name)
		}
	}
	set, _, err := command.Find([]string{"set"})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"workspace", "project"} {
		if set.Flags().Lookup(name) == nil {
			t.Fatalf("set omitted --%s", name)
		}
	}
	if factoryCalls != 0 {
		t.Fatalf("factory called during Init: %d", factoryCalls)
	}
}

func TestPromptPrintsOnlyWorkspaceAndProjectNameWithoutPlaneFactory(t *testing.T) {
	conf := &config.Config{
		Context: config.Context{
			Workspace: "my-workspace",
			Project:   config.ProjectContext{ID: "project-uuid", Name: "Project X"},
		},
		OutputFormat: "json",
		PlaneSettings: config.PlaneSettings{
			APIURL:      "https://plane.example",
			APIKey:      "secret-value",
			BearerToken: "another-secret",
		},
	}
	factoryCalls := 0
	command := Init(conf, func() (plane.Client, error) {
		factoryCalls++
		return nil, nil
	}, nil, &fakePrompt{})
	command.SetArgs([]string{"prompt"})

	var executeErr error
	output := captureStdout(t, func() {
		executeErr = command.Execute()
	})
	if executeErr != nil {
		t.Fatal(executeErr)
	}
	if output != "my-workspace | Project X\n" {
		t.Fatalf("prompt output = %q, want %q", output, "my-workspace | Project X\n")
	}
	for _, forbidden := range []string{"project-uuid", "plane.example", "secret-value", "another-secret"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("prompt output exposed %q: %q", forbidden, output)
		}
	}
	if factoryCalls != 0 {
		t.Fatalf("prompt invoked Plane factory: %d", factoryCalls)
	}
}

func TestPromptRejectsNilConfiguration(t *testing.T) {
	if err := outputPrompt(nil); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("nil prompt config error = %v", err)
	}
}

func TestSetChangedFlagsPreserveOmittedContextAndNeverPrompt(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want config.Context
	}{
		{
			name: "workspace only",
			args: []string{"--workspace", "new-workspace"},
			want: config.Context{Workspace: "new-workspace", Project: config.ProjectContext{ID: "old-id", Name: "Old Project"}},
		},
		{
			name: "project only",
			args: []string{"--project", "new-id"},
			want: config.Context{Workspace: "old-workspace", Project: config.ProjectContext{ID: "new-id", Name: "Old Project"}},
		},
		{
			name: "both",
			args: []string{"--workspace", "new-workspace", "--project", "new-id"},
			want: config.Context{Workspace: "new-workspace", Project: config.ProjectContext{ID: "new-id", Name: "Old Project"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := []string{}
			conf := &config.Config{Context: config.Context{Workspace: "old-workspace", Project: config.ProjectContext{ID: "old-id", Name: "Old Project"}}}
			prompt := &fakePrompt{events: &events}
			var saved []config.Context
			factoryCalls := 0
			command := Init(conf, func() (plane.Client, error) {
				factoryCalls++
				return &fakePlaneClient{}, nil
			}, func(value config.Context) error {
				saved = append(saved, value)
				return nil
			}, prompt)
			command.SetArgs(append([]string{"set"}, tt.args...))
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if len(saved) != 1 || saved[0] != tt.want {
				t.Fatalf("saved contexts = %#v, want %#v", saved, []config.Context{tt.want})
			}
			if len(events) != 0 || factoryCalls != 0 {
				t.Fatalf("changed flags invoked prompt/factory: events=%v factory=%d", events, factoryCalls)
			}
		})
	}
}

func TestSetRejectsBlankChangedValuesAndFailedSaveDoesNotMutateConfig(t *testing.T) {
	for _, args := range [][]string{{"--workspace", " \t"}, {"--project", " "}} {
		t.Run(strings.Join(args, "-"), func(t *testing.T) {
			conf := &config.Config{Context: config.Context{Workspace: "old-workspace", Project: config.ProjectContext{ID: "old-id", Name: "Old Project"}}}
			saveCalls := 0
			command := Init(conf, nil, func(config.Context) error {
				saveCalls++
				return errors.New("save failed")
			}, &fakePrompt{})
			command.SetArgs(append([]string{"set"}, args...))
			if err := command.Execute(); err == nil {
				t.Fatal("blank context value was accepted")
			}
			if saveCalls != 0 {
				t.Fatalf("saver called for blank value: %d", saveCalls)
			}
			if conf.Context.Workspace != "old-workspace" || conf.Context.Project.ID != "old-id" {
				t.Fatalf("config mutated for blank value: %#v", conf.Context)
			}
		})
	}

	conf := &config.Config{Context: config.Context{Workspace: "old-workspace", Project: config.ProjectContext{ID: "old-id", Name: "Old Project"}}}
	command := Init(conf, nil, func(config.Context) error { return errors.New("save failed") }, &fakePrompt{})
	command.SetArgs([]string{"set", "--workspace", "new-workspace"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "save failed") {
		t.Fatalf("failed save error = %v", err)
	}
	if conf.Context.Workspace != "old-workspace" {
		t.Fatalf("config mutated after failed save: %#v", conf.Context)
	}
}

func TestInteractiveSetPromptsWorkspaceBeforeListingAndMapsSelectedProject(t *testing.T) {
	events := []string{}
	conf := &config.Config{Context: config.Context{Project: config.ProjectContext{Name: "Old Project"}}}
	prompt := &fakePrompt{events: &events, workspace: " selected-workspace ", project: "Second Project"}
	client := &fakePlaneClient{
		events: &events,
		page: projectPage(
			projectresource.Project{ID: "first-id", Name: "First Project"},
			projectresource.Project{ID: "second-id", Name: "Second Project"},
		),
	}
	factoryCalls := 0
	var saved config.Context
	command := Init(conf, func() (plane.Client, error) {
		factoryCalls++
		events = append(events, "factory")
		return client, nil
	}, func(value config.Context) error {
		saved = value
		return nil
	}, prompt)
	if factoryCalls != 0 {
		t.Fatal("factory was called during Init")
	}
	command.SetArgs([]string{"set"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(events, ","), "workspace-prompt,factory,project-list,project-prompt"; got != want {
		t.Fatalf("event order = %s, want %s", got, want)
	}
	if len(prompt.options) != 2 || prompt.options[0] != "First Project" || prompt.options[1] != "Second Project" {
		t.Fatalf("project options = %#v", prompt.options)
	}
	want := config.Context{Workspace: "selected-workspace", Project: config.ProjectContext{ID: "second-id", Name: "Second Project"}}
	if saved != want || conf.Context != (config.Context{Project: config.ProjectContext{Name: "Old Project"}}) {
		t.Fatalf("saved=%#v conf=%#v want saved=%#v", saved, conf.Context, want)
	}
}

func TestInteractiveSetErrorsDoNotWrite(t *testing.T) {
	page := projectPage(projectresource.Project{ID: "project-id", Name: "Project"})
	tests := []struct {
		name       string
		prompt     *fakePrompt
		factoryErr error
		clientErr  error
		page       projectresource.ProjectPage
	}{
		{name: "workspace prompt", prompt: &fakePrompt{workspaceErr: errors.New("prompt failed")}},
		{name: "factory", prompt: &fakePrompt{workspace: "workspace"}, factoryErr: errors.New("factory failed")},
		{name: "list", prompt: &fakePrompt{workspace: "workspace"}, clientErr: errors.New("list failed"), page: page},
		{name: "empty", prompt: &fakePrompt{workspace: "workspace"}, page: projectPage()},
		{name: "selection", prompt: &fakePrompt{workspace: "workspace", project: "missing"}, page: page},
		{name: "selection prompt", prompt: &fakePrompt{workspace: "workspace", projectErr: errors.New("selection failed")}, page: page},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveCalls := 0
			command := Init(&config.Config{}, func() (plane.Client, error) {
				if tt.factoryErr != nil {
					return nil, tt.factoryErr
				}
				return &fakePlaneClient{page: tt.page, err: tt.clientErr}, nil
			}, func(config.Context) error {
				saveCalls++
				return nil
			}, tt.prompt)
			command.SetArgs([]string{"set"})
			if err := command.Execute(); err == nil {
				t.Fatal("interactive error was swallowed")
			}
			if saveCalls != 0 {
				t.Fatalf("saver called after interactive failure: %d", saveCalls)
			}
		})
	}
}

func TestGetUsesStandardOutputFormatsForContextOnly(t *testing.T) {
	value := config.Context{Workspace: "my-workspace", Project: config.ProjectContext{ID: "project-uuid", Name: "Project X"}}
	for _, format := range []string{"", "JSON", "YAML", "GRON", "TEXT", "TABLE", "RAW"} {
		t.Run(format, func(t *testing.T) {
			conf := &config.Config{
				Context:       value,
				OutputFormat:  format,
				NoHeaders:     true,
				PlaneSettings: config.PlaneSettings{APIURL: "https://plane.example", APIKey: "secret-value", BearerToken: "another-secret"},
			}
			output := captureStdout(t, func() {
				if err := outputContext(conf); err != nil {
					t.Fatal(err)
				}
			})
			trimmed := strings.TrimSpace(output)
			if trimmed == "" || !strings.Contains(trimmed, "my-workspace") || !strings.Contains(trimmed, "project-uuid") || !strings.Contains(trimmed, "Project X") {
				t.Fatalf("%s output = %q", format, output)
			}
			if strings.Contains(trimmed, "plane.example") || strings.Contains(trimmed, "secret-value") || strings.Contains(trimmed, "another-secret") || strings.Contains(trimmed, "api_url") {
				t.Fatalf("%s output exposed unrelated configuration: %q", format, output)
			}
			if format == "" && conf.OutputFormat != "json" {
				t.Fatalf("default output format = %q, want json", conf.OutputFormat)
			}
		})
	}
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

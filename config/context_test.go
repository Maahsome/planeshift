package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func TestContextUsesStableSerializationNames(t *testing.T) {
	value := Context{
		Workspace: "my-workspace",
		Project:   ProjectContext{ID: "project-uuid", Name: "Project X"},
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(jsonData); got != `{"workspace":"my-workspace","project":{"id":"project-uuid","name":"Project X"}}` {
		t.Fatalf("JSON = %s", got)
	}

	yamlData, err := yaml.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"workspace: my-workspace", "project:", "id: project-uuid", "name: Project X"} {
		if !strings.Contains(string(yamlData), expected) {
			t.Fatalf("YAML = %s, missing %q", yamlData, expected)
		}
	}

	settings := viper.New()
	settings.Set("workspace", "map-workspace")
	settings.Set("project.id", "map-project")
	settings.Set("project.name", "Map Project")
	var decoded Context
	if err := settings.Unmarshal(&decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Workspace != "map-workspace" || decoded.Project.ID != "map-project" || decoded.Project.Name != "Map Project" {
		t.Fatalf("mapstructure decoded = %#v", decoded)
	}
}

func TestContextZeroAndPartialValuesRemainExplicit(t *testing.T) {
	var zero Context
	if zero.Workspace != "" || zero.Project.ID != "" || zero.Project.Name != "" {
		t.Fatalf("zero context = %#v", zero)
	}

	partial := Context{Workspace: "workspace-only", Project: ProjectContext{ID: "id-only"}}
	if partial.Project.Name != "" {
		t.Fatalf("partial context changed omitted project name: %#v", partial)
	}
	data, err := json.Marshal(partial)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"workspace":"workspace-only","project":{"id":"id-only","name":""}}` {
		t.Fatalf("partial JSON = %s", data)
	}
}

func TestResolveRouteContextUsesContextDefaults(t *testing.T) {
	saved := Context{Workspace: "saved-workspace", Project: ProjectContext{ID: "saved-project"}}

	got, err := ResolveRouteContext(saved, nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if got != (RouteContext{Workspace: "saved-workspace", ProjectID: "saved-project"}) {
		t.Fatalf("resolved route context = %#v", got)
	}
}

func TestResolveRouteContextExplicitOverridesWin(t *testing.T) {
	saved := Context{Workspace: "saved-workspace", Project: ProjectContext{ID: "saved-project"}}
	workspace, projectID := "override-workspace", "override-project"

	got, err := ResolveRouteContext(saved, &workspace, &projectID, true)
	if err != nil {
		t.Fatal(err)
	}
	if got != (RouteContext{Workspace: workspace, ProjectID: projectID}) {
		t.Fatalf("resolved route context = %#v", got)
	}
	if saved.Workspace != "saved-workspace" || saved.Project.ID != "saved-project" {
		t.Fatalf("saved context was mutated: %#v", saved)
	}
}

func TestResolveRouteContextRejectsBlankOverridesAndMissingValues(t *testing.T) {
	saved := Context{Workspace: "saved-workspace", Project: ProjectContext{ID: "saved-project"}}
	blank := "   "
	for _, test := range []struct {
		name       string
		workspace  *string
		projectID  *string
		requireID  bool
		wantPhrase string
	}{
		{name: "blank workspace", workspace: &blank, projectID: nil, requireID: false, wantPhrase: "--workspace cannot be blank"},
		{name: "blank project", workspace: nil, projectID: &blank, requireID: true, wantPhrase: "--project-id cannot be blank"},
		{name: "missing workspace", workspace: nil, projectID: nil, requireID: false, wantPhrase: "context.workspace is required"},
		{name: "missing project", workspace: nil, projectID: nil, requireID: true, wantPhrase: "context.project.id is required"},
	} {
		t.Run(test.name, func(t *testing.T) {
			contextValue := saved
			if test.name == "missing workspace" {
				contextValue.Workspace = ""
			}
			if test.name == "missing project" {
				contextValue.Project.ID = ""
			}
			_, err := ResolveRouteContext(contextValue, test.workspace, test.projectID, test.requireID)
			if err == nil || !strings.Contains(err.Error(), test.wantPhrase) {
				t.Fatalf("error = %v, want phrase %q", err, test.wantPhrase)
			}
		})
	}
}

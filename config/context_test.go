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

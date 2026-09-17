package config

import (
	"fmt"
	"strings"
)

// Context contains the workspace and project selected for local CLI use.
// It intentionally contains only non-sensitive identity values, never Plane
// settings or credentials.
type Context struct {
	Workspace string         `json:"workspace" yaml:"workspace" mapstructure:"workspace"`
	Project   ProjectContext `json:"project" yaml:"project" mapstructure:"project"`
}

// ProjectContext identifies a selected Plane project.
type ProjectContext struct {
	ID   string `json:"id" yaml:"id" mapstructure:"id"`
	Name string `json:"name" yaml:"name" mapstructure:"name"`
}

// RouteContext contains the workspace and, when needed, project identity used
// to construct a resource route. It is deliberately separate from Context so
// resolving command overrides cannot mutate or persist the saved context.
type RouteContext struct {
	Workspace string
	ProjectID string
}

// ResolveRouteContext applies optional command-line overrides to the saved
// context. A non-nil override always wins, including an explicitly blank
// value, which is rejected instead of falling back to the saved value.
func ResolveRouteContext(saved Context, workspaceOverride, projectIDOverride *string, requireProject bool) (RouteContext, error) {
	workspace, err := resolveRouteValue("--workspace", workspaceOverride, saved.Workspace, "context.workspace")
	if err != nil {
		return RouteContext{}, err
	}
	projectID := ""
	if requireProject || projectIDOverride != nil {
		projectID, err = resolveRouteValue("--project-id", projectIDOverride, saved.Project.ID, "context.project.id")
		if err != nil {
			return RouteContext{}, err
		}
	}
	return RouteContext{Workspace: workspace, ProjectID: projectID}, nil
}

func resolveRouteValue(flag string, override *string, fallback, contextKey string) (string, error) {
	if override != nil {
		if strings.TrimSpace(*override) == "" {
			return "", fmt.Errorf("%s cannot be blank", flag)
		}
		return *override, nil
	}
	if strings.TrimSpace(fallback) == "" {
		return "", fmt.Errorf("%s is required", contextKey)
	}
	return fallback, nil
}

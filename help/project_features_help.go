package help

// ProjectFeaturesCmd provides safe help for the nested Project Features
// resource. Examples use placeholders that contain no credentials or
// sensitive Plane data.
type ProjectFeaturesCmd struct{}

func (p *ProjectFeaturesCmd) Short() string {
	return "Manage project feature flags"
}

func (p *ProjectFeaturesCmd) Long() string {
	return `Manage feature flags for one project.

Use a workspace slug first and the project ID second:
  planeshift project features get my-workspace project-uuid

Available operations:
  get workspace_slug project_id
  update workspace_slug project_id [flags]

The update operation is partial. Each optional boolean flag is sent only when
you provide it, and an explicit --flag=false preserves false on the request.
Supported update flags:
  --epics, --modules, --cycles, --views, --pages, --intakes, --work-item-types

Examples:
  planeshift project features update my-workspace project-uuid --cycles=false
  planeshift project features update my-workspace project-uuid --pages --views

Use the root -o/--output option to select json, yaml, gron, text, table, or raw output.`
}

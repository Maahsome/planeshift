package help

// ProjectsCmd provides safe, resource-specific help for the Projects command
// surface. Examples intentionally contain identifiers only, never credentials
// or upload/presigned data.
type ProjectsCmd struct{}

func (p *ProjectsCmd) Short() string {
	return "Manage Plane projects"
}

func (p *ProjectsCmd) Long() string {
	return `Manage the projects in a workspace.

Project commands use a workspace slug as their first positional argument. Commands that address one project take the project ID second. List uses cursor pagination and defaults to Plane's server page size of 20.

Available operations:
  list, create, create-template, get, update, archive, unarchive, delete

Examples:
  planeshift get projects list my-workspace --per-page 20
  planeshift get projects create my-workspace --name "Project X" --identifier PROJX
  planeshift get projects archive my-workspace project-uuid

The archive and unarchive operations return 204 with no output. Use --icon-prop with a JSON object or value when a project icon is needed; dynamic response fields are retained by JSON/YAML/GRON/raw output.`
}

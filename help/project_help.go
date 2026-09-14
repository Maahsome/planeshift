package help

// ProjectCmd provides safe, resource-specific help for the Project command
// surface. Examples intentionally contain identifiers only, never credentials
// or upload/presigned data.
type ProjectCmd struct{}

func (p *ProjectCmd) Short() string {
	return "Manage Plane projects"
}

func (p *ProjectCmd) Long() string {
	return `Manage projects in a workspace.

Use the singular command name for the canonical invocation; projects is an alias.
Project commands use a workspace slug as their first positional argument. Commands that address one project take the project ID second. List uses cursor pagination and defaults to Plane's server page size of 20.

Available operations:
  list, create, create-template, get, update, archive, unarchive, delete, features

Examples:
  planeshift project list my-workspace --per-page 20
  planeshift project create my-workspace --name "Project X" --identifier PROJX
  planeshift project archive my-workspace project-uuid
  planeshift projects list my-workspace

The archive and unarchive operations return 204 with no output. Use --icon-prop with a JSON object or value when a project icon is needed; dynamic response fields are retained by JSON/YAML/GRON/raw output.`
}

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
Project commands take route values from the saved context by default. Use
--workspace to override context.workspace. Commands that address one project
also accept --project-id to override context.project.id. An explicit override
wins; an explicitly blank value is rejected. There are no positional route
arguments. List uses cursor pagination and defaults to Plane's server page size
of 20.

Available operations:
  list, create, create-template, get, update, archive, unarchive, delete

Examples:
  planeshift project list --per-page 20
  planeshift project create --workspace my-workspace --name "Project X" --identifier PROJX
  planeshift project archive --workspace my-workspace --project-id project-uuid
  planeshift projects list --workspace my-workspace

The archive and unarchive operations return 204 with no output. Use --icon-prop with a JSON object or value when a project icon is needed; dynamic response fields are retained by JSON/YAML/GRON/raw output.`
}

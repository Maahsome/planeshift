package help

// ContextCmd provides safe help for the local workspace/project context.
type ContextCmd struct{}

// ContextSetCmd provides help for context set.
type ContextSetCmd struct{}

// ContextGetCmd provides help for context get.
type ContextGetCmd struct{}

func (c *ContextCmd) Short() string { return "Manage the local workspace and project context" }

func (c *ContextCmd) Long() string {
	return `Manage the workspace and project selected for local planeshift use.

Use context set with --workspace and/or --project for non-interactive updates.
Omitted values remain unchanged; --project stores a project ID and preserves
the previously stored project name. With no flags, set prompts for the
workspace first, then lists projects and lets you select by project name.

The context is stored as context.workspace, context.project.id, and
context.project.name. context get prints only those values.

Inherited output options include --output (json, text, yaml, gron, or raw) and
--no-headers for text output.

Examples:
  planeshift context set --workspace my-workspace
  planeshift context set --project project-uuid
  planeshift context get --output yaml`
}

func (c *ContextSetCmd) Short() string { return "Set the current workspace and project context" }

func (c *ContextSetCmd) Long() string {
	return `Set the local context.

Pass --workspace and/or --project for a non-interactive update. Omitted values
remain unchanged. With no flags, the command prompts for the workspace first,
then lists projects and lets you select one by name. The selected workspace,
project ID, and project name are saved together.

Inherited output options are available for commands that produce output.`
}

func (c *ContextGetCmd) Short() string { return "Show the current workspace and project context" }

func (c *ContextGetCmd) Long() string {
	return `Show only the local context: workspace, project ID, and project name.

Output uses the inherited --output option and supports JSON, YAML, GRON, text,
table, and raw formats. Use --no-headers with text output when needed.`
}

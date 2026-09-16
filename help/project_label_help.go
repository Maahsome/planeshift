package help

// ProjectLabelCmd provides safe help for the workspace-scoped Project Labels
// command. It intentionally contains placeholders only, never credentials or
// sensitive Plane response data.
type ProjectLabelCmd struct{}

func (p *ProjectLabelCmd) Short() string {
	return "Manage workspace project labels"
}

func (p *ProjectLabelCmd) Long() string {
	return `Manage reusable project labels in a workspace.

Use project-label as the canonical command name; project-labels is an alias.
Every operation takes workspace_slug first. Get, update, and delete take
label_id second. Project labels are workspace-scoped and are not nested under
a project ID.

Available operations:
  create workspace_slug
  list workspace_slug
  get workspace_slug label_id
  update workspace_slug label_id
  delete workspace_slug label_id

List supports --cursor, --per-page (default 20, maximum 100), --fields,
--expand, and --order-by. Create requires --name and also accepts
--description, --color, and --sort-order. Update is partial: untouched fields
are omitted, while explicit empty strings and --sort-order 0 are sent.

Delete returns 204 with no output. Use the root -o/--output option for json,
yaml, gron, text, table, or raw output.`
}

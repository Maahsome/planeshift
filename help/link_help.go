package help

// LinkCmd provides safe, resource-specific help for Work Item Links.
type LinkCmd struct{}

func (l *LinkCmd) Short() string { return "Manage Plane Work Item Links" }

func (l *LinkCmd) Long() string {
	return `Manage external links attached to work items.

The canonical command is link; links is its plural alias. Project-scoped
commands use --workspace and --project-id, defaulting to context.workspace and
context.project.id. Explicit route flags override saved context values; blank
overrides and missing required context values fail before a request.

Operations:
  list work_item_id, create work_item_id, get work_item_id link_id
  update work_item_id link_id, delete work_item_id link_id

List supports --cursor, --per-page, --fields, --expand, and --order-by. The
detail get operation supports --cursor, --per-page, --fields, and --expand.
Create requires --url and accepts --title. Update accepts only the optional
--url and --title flags; omitted values are unchanged, including explicit
empty strings when a flag is supplied. Delete returns 204 with no output.
JSON, YAML, GRON, text, table, and raw output use the configured output path.

The hidden link legacy subtree contains only the five explicitly inventoried
/issues/ compatibility operations and is not an issue command or an alias.

Examples:
  planeshift link list work-item-uuid --per-page 20
  planeshift links create work-item-uuid --url https://example.com --title "Design"
  planeshift link update work-item-uuid link-uuid --title "Updated"
  planeshift link delete work-item-uuid link-uuid`
}

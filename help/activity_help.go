package help

// ActivityCmd provides safe, resource-specific help for Work Item Activity.
type ActivityCmd struct{}

func (a *ActivityCmd) Short() string { return "Read Plane Work Item Activity history" }

func (a *ActivityCmd) Long() string {
	return `Read the change history attached to Plane work items.

The canonical command is activity; activities is its plural alias. Project-
scoped commands use --workspace and --project-id, defaulting to
context.workspace and context.project.id. Explicit route flags override saved
context values; blank overrides and missing required context values fail before
a request.

Operations:
  list work_item_id, get work_item_id activity_id

Both operations support --cursor, --per-page, --fields, --expand, and
--order-by. Results use the configured JSON, YAML, GRON, text, table, or raw
output path. Activity reads are read-only; creation, updates, and deletes are
not available.

The hidden activity legacy subtree contains only the two explicitly inventoried
/issues/ compatibility reads: list issue_id and get issue_id activity_id. It
is compatibility-only and is not an issue command or an additional alias.

Examples:
  planeshift activity list work-item-uuid --per-page 20
  planeshift activities get work-item-uuid activity-uuid --expand actor`
}

package help

// LabelCmd provides safe, resource-specific help for project-scoped labels.
type LabelCmd struct{}

func (l *LabelCmd) Short() string { return "Manage project-scoped Plane labels" }

func (l *LabelCmd) Long() string {
	return `Manage labels in a project.

The canonical command is label; labels is its plural alias. Every operation
uses --workspace and --project-id, defaulting to context.workspace and
context.project.id. Explicit route flags override the saved context; blank
overrides and missing required context values fail before a request. Get,
update, and delete retain only the label ID as a positional argument.

Operations:
  list, create, get, update, delete

List supports only --cursor, --per-page, --fields, --expand, and --order-by.
The server defaults to 20 results and accepts page sizes from 1 through 100.
Create requires --name and accepts color, description, external-source,
external-id, parent, and sort-order. Update is partial; omitted fields are
unchanged, while explicitly supplied empty strings and zero are sent.
Create returns 201; list, get, and update return 200.

Delete succeeds with 204 and prints no output. JSON, YAML, GRON, text, table,
and raw output use the configured centralized output path.

Examples:
  planeshift label list --per-page 20 --order-by name
  planeshift labels create --workspace my-workspace --project-id project-uuid --name "Urgent"
  planeshift label update --project-id project-uuid label-uuid --sort-order 0
  planeshift label delete --workspace my-workspace --project-id project-uuid label-uuid`
}

package help

// WorkItemCmd provides safe help for the primary Work Item resource and its
// explicitly separated legacy compatibility subtree.
type WorkItemCmd struct{}

func (w *WorkItemCmd) Short() string { return "Manage Plane work items" }

func (w *WorkItemCmd) Long() string {
	return `Manage work items in a project.

The canonical command is work-item; work-items is its plural alias. Primary
commands use the current /work-items/ API family. Project-scoped commands use
--workspace and --project-id, defaulting to context.workspace and
context.project.id. Search and identifier lookup use --workspace and do not
require a project route value. Explicit route flags override context values;
blank overrides and missing required context values fail before a request.
Identifier lookup retains project_identifier and issue_identifier as
positional arguments because they are human identifiers required by the API.

Operations:
  search, get-by-identifier, list, create, get, update, delete
  relations-list, relations-create

List supports cursor, per-page, fields, expand, external-id, external-source,
and order-by. Search requires --search and supports limit, project-id, and
workspace-search. Search's --project-id is only a query filter; it is not the
project route override used by project-scoped commands. Detail reads support expand, fields, external-id,
external-source, and order-by. Create requires --name; update sends only
selected partial fields, including explicit false, zero, empty arrays, or
JSON null when constructed through the request contract.

Relation creation accepts blocking, blocked_by, duplicate, relates_to,
start_before, start_after, finish_before, and finish_after with related IDs.
JSON, YAML, GRON, text, table, and raw output use the configured output path.
Successful deletes return 204 with no output.

The hidden legacy subtree is available only as work-item legacy and contains
the seven explicitly inventoried core /issues/ compatibility routes. It does
not include relation commands or any other deprecated route family.

Examples:
  planeshift work-item search --workspace my-workspace --search "release"
  planeshift work-item list --workspace my-workspace --project-id project-uuid
  planeshift work-items get-by-identifier --workspace my-workspace ENG 123
  planeshift work-item create --project-id project-uuid --name "Document API"
  planeshift work-item relations-create --project-id project-uuid work-item-uuid --relation-type relates_to --issue related-item-uuid`
}

package help

// WorkItemCmd provides safe help for the primary Work Item resource and its
// explicitly separated legacy compatibility subtree.
type WorkItemCmd struct{}

func (w *WorkItemCmd) Short() string { return "Manage Plane work items" }

func (w *WorkItemCmd) Long() string {
	return `Manage work items in a project.

The canonical command is work-item; work-items is its plural alias. Primary
commands use the current /work-items/ API family. Positional arguments start
with the workspace slug, followed by the project ID where a project scope is
needed. Identifier lookup takes workspace slug, project identifier, and issue
identifier in that order.

Operations:
  search, get-by-identifier, list, create, get, update, delete
  relations-list, relations-create

List supports cursor, per-page, fields, expand, external-id, external-source,
and order-by. Search requires --search and supports limit, project-id, and
workspace-search. Detail reads support expand, fields, external-id,
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
  planeshift work-item search my-workspace --search "release"
  planeshift work-item list my-workspace 00000000-0000-0000-0000-000000000001
  planeshift work-items get-by-identifier my-workspace ENG 123
  planeshift work-item create my-workspace 00000000-0000-0000-0000-000000000001 --name "Document API"
  planeshift work-item relations-create my-workspace 00000000-0000-0000-0000-000000000001 00000000-0000-0000-0000-000000000002 --relation-type relates_to --issue 00000000-0000-0000-0000-000000000003`
}

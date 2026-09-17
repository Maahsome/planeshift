package help

// StateCmd provides safe, resource-specific help for Work Item States.
type StateCmd struct{}

func (s *StateCmd) Short() string { return "Manage Plane Work Item States" }

func (s *StateCmd) Long() string {
	return `Manage workflow states in a project.

The canonical command is state; states is its plural alias. Every operation
starts with a workspace slug and project ID. Get, update, and delete also take
a state ID.

Operations:
  list, create, get, update, delete

List supports only --cursor, --per-page, --fields, and --expand. The server
defaults to 20 results and accepts page sizes from 1 through 100. Create
requires --name and --color and accepts description, sequence, group,
is-triage, default, external-source, and external-id. Supported groups include
backlog, unstarted, started, completed, cancelled, and triage. Sequence accepts
a JSON number, string, or null. Update is partial; omitted flags are unchanged,
while explicitly supplied empty, false, zero, and null values are sent.

Delete succeeds with 204 and prints no output. JSON, YAML, GRON, text, table,
and raw output use the configured centralized output path.

Examples:
  planeshift state list my-workspace project-uuid --per-page 20
  planeshift states create my-workspace project-uuid --name "Started" --color "#00ff00"
  planeshift state update my-workspace project-uuid state-uuid --sequence 2
  planeshift state delete my-workspace project-uuid state-uuid`
}

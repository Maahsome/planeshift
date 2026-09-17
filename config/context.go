package config

// Context contains the workspace and project selected for local CLI use.
// It intentionally contains only non-sensitive identity values, never Plane
// settings or credentials.
type Context struct {
	Workspace string         `json:"workspace" yaml:"workspace" mapstructure:"workspace"`
	Project   ProjectContext `json:"project" yaml:"project" mapstructure:"project"`
}

// ProjectContext identifies a selected Plane project.
type ProjectContext struct {
	ID   string `json:"id" yaml:"id" mapstructure:"id"`
	Name string `json:"name" yaml:"name" mapstructure:"name"`
}

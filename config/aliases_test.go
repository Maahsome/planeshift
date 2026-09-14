package config

import "testing"

func TestProjectCommandNames(t *testing.T) {
	if ProjectCommandName != "project" {
		t.Fatalf("ProjectCommandName = %q, want project", ProjectCommandName)
	}
	if ProjectCommandAlias != "projects" {
		t.Fatalf("ProjectCommandAlias = %q, want projects", ProjectCommandAlias)
	}
}

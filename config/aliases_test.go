package config

import "testing"

func TestProjectCommandNames(t *testing.T) {
	if ProjectCommandName != "project" {
		t.Fatalf("ProjectCommandName = %q, want project", ProjectCommandName)
	}
	if ProjectCommandAlias != "projects" {
		t.Fatalf("ProjectCommandAlias = %q, want projects", ProjectCommandAlias)
	}
	if ProjectLabelCommandName != "project-label" {
		t.Fatalf("ProjectLabelCommandName = %q, want project-label", ProjectLabelCommandName)
	}
	if ProjectLabelCommandAlias != "project-labels" {
		t.Fatalf("ProjectLabelCommandAlias = %q, want project-labels", ProjectLabelCommandAlias)
	}
}

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

func TestWorkItemCommandNames(t *testing.T) {
	if WorkItemCommandName != "work-item" {
		t.Fatalf("WorkItemCommandName = %q, want work-item", WorkItemCommandName)
	}
	if WorkItemCommandAlias != "work-items" {
		t.Fatalf("WorkItemCommandAlias = %q, want work-items", WorkItemCommandAlias)
	}
}

func TestStateCommandNames(t *testing.T) {
	if StateCommandName != "state" {
		t.Fatalf("StateCommandName = %q, want state", StateCommandName)
	}
	if StateCommandAlias != "states" {
		t.Fatalf("StateCommandAlias = %q, want states", StateCommandAlias)
	}
}

func TestContextCommandName(t *testing.T) {
	if ContextCommandName != "context" {
		t.Fatalf("ContextCommandName = %q, want context", ContextCommandName)
	}
}

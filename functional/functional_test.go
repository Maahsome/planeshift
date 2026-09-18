package functional

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// commandSpec is the black-box inventory for the CLI surface covered by this
// package. The usage text is asserted from Cobra help rather than importing
// command packages, keeping the functional suite on the executable boundary.
type commandSpec struct {
	path          []string
	usage         string
	flags         []string
	requiredFlags []string
}

var commandCoverage = []commandSpec{
	{path: []string{"version"}, usage: "version"},
	{path: []string{"context"}, usage: "context"},
	{path: []string{"context", "set"}, usage: "set", requiredFlags: []string{"--workspace", "--project"}},
	{path: []string{"context", "get"}, usage: "get"},
	{path: []string{"context", "prompt"}, usage: "prompt"},
	{path: []string{"project"}, usage: "project"},
	{path: []string{"projects"}, usage: "project"},
	{path: []string{"project", "list"}, usage: "list", flags: []string{"--workspace"}},
	{path: []string{"projects", "list"}, usage: "list", flags: []string{"--workspace"}},
	{path: []string{"project", "create"}, usage: "create", flags: []string{"--workspace"}, requiredFlags: []string{"--name", "--identifier"}},
	{path: []string{"project", "create-template"}, usage: "create-template", flags: []string{"--workspace"}, requiredFlags: []string{"--template-id"}},
	{path: []string{"project", "get"}, usage: "get", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"project", "update"}, usage: "update", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"project", "archive"}, usage: "archive", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"project", "unarchive"}, usage: "unarchive", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"project", "delete"}, usage: "delete", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"work-item"}, usage: "work-item"},
	{path: []string{"work-items"}, usage: "work-item"},
	{path: []string{"work-item", "search"}, usage: "search", flags: []string{"--workspace", "--project-id"}, requiredFlags: []string{"--search"}},
	{path: []string{"work-item", "get-by-identifier"}, usage: "get-by-identifier project_identifier issue_identifier", flags: []string{"--workspace"}},
	{path: []string{"work-item", "list"}, usage: "list", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"work-items", "list"}, usage: "list", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"work-item", "create"}, usage: "create", flags: []string{"--workspace", "--project-id"}, requiredFlags: []string{"--name"}},
	{path: []string{"work-item", "get"}, usage: "get work_item_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"work-item", "update"}, usage: "update work_item_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"work-item", "delete"}, usage: "delete work_item_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"work-item", "relations-list"}, usage: "relations-list work_item_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"work-item", "relations-create"}, usage: "relations-create work_item_id", flags: []string{"--workspace", "--project-id"}, requiredFlags: []string{"--relation-type", "--issue"}},
	{path: []string{"work-item", "legacy"}, usage: "legacy"},
	{path: []string{"work-item", "legacy", "search"}, usage: "search", flags: []string{"--workspace", "--project-id"}, requiredFlags: []string{"--search"}},
	{path: []string{"work-item", "legacy", "get-by-identifier"}, usage: "get-by-identifier project_identifier issue_identifier", flags: []string{"--workspace"}},
	{path: []string{"work-item", "legacy", "list"}, usage: "list", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"work-item", "legacy", "create"}, usage: "create", flags: []string{"--workspace", "--project-id"}, requiredFlags: []string{"--name"}},
	{path: []string{"work-item", "legacy", "get"}, usage: "get work_item_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"work-item", "legacy", "update"}, usage: "update work_item_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"work-item", "legacy", "delete"}, usage: "delete work_item_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"state"}, usage: "state"},
	{path: []string{"states"}, usage: "state"},
	{path: []string{"state", "list"}, usage: "list", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"states", "list"}, usage: "list", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"state", "create"}, usage: "create", flags: []string{"--workspace", "--project-id"}, requiredFlags: []string{"--name", "--color"}},
	{path: []string{"state", "get"}, usage: "get state_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"state", "update"}, usage: "update state_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"state", "delete"}, usage: "delete state_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"label"}, usage: "label"},
	{path: []string{"labels"}, usage: "label"},
	{path: []string{"label", "list"}, usage: "list", flags: []string{"--workspace", "--project-id", "--cursor", "--per-page", "--fields", "--expand", "--order-by"}},
	{path: []string{"labels", "list"}, usage: "list", flags: []string{"--workspace", "--project-id", "--cursor", "--per-page", "--fields", "--expand", "--order-by"}},
	{path: []string{"label", "create"}, usage: "create", flags: []string{"--workspace", "--project-id", "--color", "--description", "--external-source", "--external-id", "--parent", "--sort-order"}, requiredFlags: []string{"--name"}},
	{path: []string{"label", "get"}, usage: "get label_id", flags: []string{"--workspace", "--project-id"}},
	{path: []string{"label", "update"}, usage: "update label_id", flags: []string{"--workspace", "--project-id", "--name", "--color", "--description", "--external-source", "--external-id", "--parent", "--sort-order"}},
	{path: []string{"label", "delete"}, usage: "delete label_id", flags: []string{"--workspace", "--project-id"}},
}

func TestCommandCoverageMatrix(t *testing.T) {
	runner, ok := discoveryRunner(t)
	if !ok {
		return
	}

	for _, spec := range commandCoverage {
		spec := spec
		t.Run(strings.Join(spec.path, "/"), func(t *testing.T) {
			args := append(append([]string(nil), spec.path...), "--help")
			result := runner.Run(t, args...)
			result.RequireSuccess(t)
			help := result.Output()
			if !strings.Contains(help, spec.usage) {
				t.Fatalf("help for %s omitted usage %q:\n%s", strings.Join(spec.path, " "), spec.usage, help)
			}
			for _, flag := range spec.requiredFlags {
				if !strings.Contains(help, flag) {
					t.Fatalf("help for %s omitted required flag %q:\n%s", strings.Join(spec.path, " "), flag, help)
				}
			}
			for _, flag := range spec.flags {
				if !strings.Contains(help, flag) {
					t.Fatalf("help for %s omitted context flag %q:\n%s", strings.Join(spec.path, " "), flag, help)
				}
			}
			assertNoCredentialMaterial(t, help)
		})
	}
}

func TestVersionAndRootHelpAreSafe(t *testing.T) {
	runner, ok := discoveryRunner(t)
	if !ok {
		return
	}

	version := runner.Run(t, "version", "--output", "json")
	version.RequireSuccess(t)
	if !json.Valid(bytes.TrimSpace([]byte(version.Stdout))) {
		t.Fatalf("version did not return valid JSON: %s", version.Output())
	}
	if object := jsonObject(json.RawMessage(bytes.TrimSpace([]byte(version.Stdout)))); object == nil || object["SemVer"] == nil {
		t.Fatalf("version JSON omitted SemVer: %s", version.Output())
	}
	assertNoCredentialMaterial(t, version.Output())

	for _, path := range [][]string{
		{}, {"context"}, {"context", "set"}, {"context", "get"}, {"context", "prompt"}, {"project"}, {"projects"}, {"work-item"}, {"work-items"}, {"work-item", "legacy"}, {"state"}, {"states"}, {"label"}, {"labels"},
	} {
		args := append(append([]string(nil), path...), "--help")
		result := runner.Run(t, args...)
		result.RequireSuccess(t)
		assertNoCredentialMaterial(t, result.Output())
	}
}

type managedProject struct {
	runner    *Runner
	workspace string
	id        string
	archived  bool
	deleted   bool
}

func registerProjectCleanup(t *testing.T, runner *Runner, workspace, id string) *managedProject {
	t.Helper()
	project := &managedProject{runner: runner, workspace: workspace, id: id}
	t.Cleanup(func() {
		if project.deleted {
			t.Logf("functional cleanup: project %s was already deleted by the lifecycle", project.id)
			return
		}
		if project.archived {
			result := runner.Run(t, "project", "unarchive", "--workspace", workspace, "--project-id", id)
			if result.Err != nil || result.ExitCode != 0 {
				t.Logf("functional cleanup: could not unarchive generated project %s: %s", id, result.Output())
			} else {
				project.archived = false
			}
		}
		result := runner.Run(t, "project", "delete", "--workspace", workspace, "--project-id", id)
		if result.Err != nil || result.ExitCode != 0 {
			t.Logf("functional cleanup: generated project %s was already deleted or could not be deleted: %s", id, result.Output())
		}
	})
	return project
}

type managedState struct {
	runner    *Runner
	workspace string
	projectID string
	id        string
	deleted   bool
}

func registerStateCleanup(t *testing.T, runner *Runner, workspace, projectID, id string) *managedState {
	t.Helper()
	state := &managedState{runner: runner, workspace: workspace, projectID: projectID, id: id}
	t.Cleanup(func() {
		if state.deleted {
			t.Logf("functional cleanup: state %s was already deleted by the lifecycle", state.id)
			return
		}
		result := runner.Run(t, "state", "delete", "--workspace", workspace, "--project-id", projectID, id)
		if result.Err != nil || result.ExitCode != 0 {
			t.Logf("functional cleanup: generated state %s was already deleted or could not be deleted: %s", id, result.Output())
		}
	})
	return state
}

type managedLabel struct {
	runner    *Runner
	workspace string
	projectID string
	id        string
	deleted   bool
}

func registerLabelCleanup(t *testing.T, runner *Runner, workspace, projectID, id string) *managedLabel {
	t.Helper()
	label := &managedLabel{runner: runner, workspace: workspace, projectID: projectID, id: id}
	t.Cleanup(func() {
		if label.deleted {
			t.Logf("functional cleanup: label %s was already deleted by the lifecycle", label.id)
			return
		}
		result := runner.Run(t, "label", "delete", "--workspace", workspace, "--project-id", projectID, id)
		if result.Err != nil || result.ExitCode != 0 {
			t.Logf("functional cleanup: generated label %s was already deleted or could not be deleted: %s", id, result.Output())
		}
	})
	return label
}

func TestLabelLifecycle(t *testing.T) {
	config, ok := lifecycleConfig(t)
	if !ok {
		return
	}
	runner, err := newRunner(t, config)
	if err != nil {
		t.Fatalf("configure functional runner: %v", err)
	}

	projectName := uniqueName("labels-project")
	projectData, _ := runner.RunJSON(t, "project", "create", "--workspace", config.WorkspaceSlug,
		"--name", projectName, "--identifier", uniqueIdentifier())
	project := registerProjectCleanup(t, runner, config.WorkspaceSlug, projectID(t, projectData))

	labelName := uniqueName("label")
	created, _ := runner.RunJSON(t, "label", "create", "--workspace", config.WorkspaceSlug, "--project-id", project.id,
		"--name", labelName, "--color", "#123456")
	managed := registerLabelCleanup(t, runner, config.WorkspaceSlug, project.id, labelID(t, created))
	if !jsonContainsString(created, managed.id) || !jsonContainsString(created, labelName) {
		t.Fatalf("created label omitted generated label %s/%q: %s", managed.id, labelName, created)
	}

	list, _ := runner.RunJSON(t, "label", "list", "--workspace", config.WorkspaceSlug, "--project-id", project.id, "--per-page", "20")
	if !jsonContainsString(list, managed.id) || !jsonContainsString(list, labelName) {
		t.Fatalf("label list omitted generated label %s/%q: %s", managed.id, labelName, list)
	}
	aliasList, _ := runner.RunJSON(t, "labels", "list", "--workspace", config.WorkspaceSlug, "--project-id", project.id)
	if !jsonContainsString(aliasList, managed.id) || !jsonContainsString(aliasList, labelName) {
		t.Fatalf("labels list alias omitted generated label %s/%q: %s", managed.id, labelName, aliasList)
	}

	got, _ := runner.RunJSON(t, "label", "get", "--workspace", config.WorkspaceSlug, "--project-id", project.id, managed.id)
	if !jsonContainsString(got, managed.id) || !jsonContainsString(got, labelName) {
		t.Fatalf("label get did not return generated label %s/%q: %s", managed.id, labelName, got)
	}

	updatedName := uniqueName("label-updated")
	updated, _ := runner.RunJSON(t, "label", "update", "--workspace", config.WorkspaceSlug, "--project-id", project.id, managed.id, "--name", updatedName)
	if !jsonContainsString(updated, managed.id) || !jsonContainsString(updated, updatedName) {
		t.Fatalf("label update did not return generated label %s with updated name %q: %s", managed.id, updatedName, updated)
	}
	persisted, _ := runner.RunJSON(t, "label", "get", "--workspace", config.WorkspaceSlug, "--project-id", project.id, managed.id)
	if !jsonContainsString(persisted, managed.id) || !jsonContainsString(persisted, updatedName) {
		t.Fatalf("label get did not persist updated name %q for %s: %s", updatedName, managed.id, persisted)
	}

	runner.Run(t, "label", "delete", "--workspace", config.WorkspaceSlug, "--project-id", project.id, managed.id).RequireQuietSuccess(t)
	managed.deleted = true
}

func TestProjectLifecycle(t *testing.T) {
	config, ok := lifecycleConfig(t)
	if !ok {
		return
	}
	runner, err := newRunner(t, config)
	if err != nil {
		t.Fatalf("configure functional runner: %v", err)
	}

	list, _ := runner.RunJSON(t, "project", "list", "--workspace", config.WorkspaceSlug)
	if len(list) == 0 {
		t.Fatal("project list returned no JSON document")
	}
	aliasList, _ := runner.RunJSON(t, "projects", "list", "--workspace", config.WorkspaceSlug)
	if len(aliasList) == 0 {
		t.Fatal("projects list alias returned no JSON document")
	}

	name := uniqueName("project")
	identifier := uniqueIdentifier()
	created, _ := runner.RunJSON(t, "project", "create", "--workspace", config.WorkspaceSlug,
		"--name", name, "--identifier", identifier)
	project := registerProjectCleanup(t, runner, config.WorkspaceSlug, projectID(t, created))
	if got := projectIdentifier(t, created); got != identifier {
		t.Fatalf("created project identifier = %q, want %q", got, identifier)
	}
	if !jsonContainsString(created, name) {
		t.Fatalf("created project response omitted generated name %q: %s", name, created)
	}

	got, _ := runner.RunJSON(t, "project", "get", "--workspace", config.WorkspaceSlug, "--project-id", project.id)
	if !jsonContainsString(got, project.id) || !jsonContainsString(got, name) {
		t.Fatalf("project get did not return generated project %s: %s", project.id, got)
	}

	updatedName := uniqueName("project-updated")
	updated, _ := runner.RunJSON(t, "project", "update", "--workspace", config.WorkspaceSlug, "--project-id", project.id, "--name", updatedName)
	if !jsonContainsString(updated, updatedName) {
		t.Fatalf("project update omitted updated name %q: %s", updatedName, updated)
	}
	listed, _ := runner.RunJSON(t, "project", "list", "--workspace", config.WorkspaceSlug)
	if !jsonContainsString(listed, project.id) && !jsonContainsString(listed, updatedName) {
		t.Fatalf("project list omitted generated project %s/%q: %s", project.id, updatedName, listed)
	}

	runner.Run(t, "project", "archive", "--workspace", config.WorkspaceSlug, "--project-id", project.id).RequireQuietSuccess(t)
	project.archived = true
	runner.Run(t, "project", "unarchive", "--workspace", config.WorkspaceSlug, "--project-id", project.id).RequireQuietSuccess(t)
	project.archived = false
	runner.Run(t, "project", "delete", "--workspace", config.WorkspaceSlug, "--project-id", project.id).RequireQuietSuccess(t)
	project.deleted = true

	if config.TemplateID == "" {
		t.Log("coverage status: project create-template registered/help-tested; execution skipped because PLANE_FUNCTIONAL_PROJECT_TEMPLATE_ID is not set")
		return
	}
	templateName := uniqueName("template-project")
	templateIdentifier := uniqueIdentifier()
	template, _ := runner.RunJSON(t, "project", "create-template", "--workspace", config.WorkspaceSlug,
		"--template-id", config.TemplateID, "--name", templateName, "--identifier", templateIdentifier)
	templateProject := registerProjectCleanup(t, runner, config.WorkspaceSlug, projectID(t, template))
	if got := projectIdentifier(t, template); got != templateIdentifier {
		t.Fatalf("template project identifier = %q, want %q", got, templateIdentifier)
	}
	runner.Run(t, "project", "archive", "--workspace", config.WorkspaceSlug, "--project-id", templateProject.id).RequireQuietSuccess(t)
	templateProject.archived = true
	runner.Run(t, "project", "unarchive", "--workspace", config.WorkspaceSlug, "--project-id", templateProject.id).RequireQuietSuccess(t)
	templateProject.archived = false
	runner.Run(t, "project", "delete", "--workspace", config.WorkspaceSlug, "--project-id", templateProject.id).RequireQuietSuccess(t)
	templateProject.deleted = true
}

func TestStateLifecycle(t *testing.T) {
	config, ok := lifecycleConfig(t)
	if !ok {
		return
	}
	runner, err := newRunner(t, config)
	if err != nil {
		t.Fatalf("configure functional runner: %v", err)
	}

	projectData, _ := runner.RunJSON(t, "project", "create", "--workspace", config.WorkspaceSlug,
		"--name", uniqueName("states-project"), "--identifier", uniqueIdentifier())
	project := registerProjectCleanup(t, runner, config.WorkspaceSlug, projectID(t, projectData))

	stateName := uniqueName("state")
	created, _ := runner.RunJSON(t, "state", "create", "--workspace", config.WorkspaceSlug, "--project-id", project.id,
		"--name", stateName, "--color", "#123456", "--group", "started", "--default=false", "--is-triage=false")
	managed := registerStateCleanup(t, runner, config.WorkspaceSlug, project.id, stateID(t, created))
	if !jsonContainsString(created, stateName) {
		t.Fatalf("created state response omitted generated name %q: %s", stateName, created)
	}

	list, _ := runner.RunJSON(t, "state", "list", "--workspace", config.WorkspaceSlug, "--project-id", project.id, "--per-page", "20")
	if !jsonContainsString(list, managed.id) || !jsonContainsString(list, stateName) {
		t.Fatalf("state list omitted generated state %s/%q: %s", managed.id, stateName, list)
	}
	aliasList, _ := runner.RunJSON(t, "states", "list", "--workspace", config.WorkspaceSlug, "--project-id", project.id)
	if !jsonContainsString(aliasList, managed.id) || !jsonContainsString(aliasList, stateName) {
		t.Fatalf("states list alias omitted generated state %s/%q: %s", managed.id, stateName, aliasList)
	}

	got, _ := runner.RunJSON(t, "state", "get", "--workspace", config.WorkspaceSlug, "--project-id", project.id, managed.id)
	if !jsonContainsString(got, managed.id) || !jsonContainsString(got, stateName) {
		t.Fatalf("state get did not return generated state %s/%q: %s", managed.id, stateName, got)
	}

	updatedName := uniqueName("state-updated")
	updated, _ := runner.RunJSON(t, "state", "update", "--workspace", config.WorkspaceSlug, "--project-id", project.id, managed.id, "--name", updatedName)
	if !jsonContainsString(updated, managed.id) || !jsonContainsString(updated, updatedName) {
		t.Fatalf("state update did not return generated state %s with updated name %q: %s", managed.id, updatedName, updated)
	}
	persisted, _ := runner.RunJSON(t, "state", "get", "--workspace", config.WorkspaceSlug, "--project-id", project.id, managed.id)
	if !jsonContainsString(persisted, managed.id) || !jsonContainsString(persisted, updatedName) {
		t.Fatalf("state get did not persist updated name %q for %s: %s", updatedName, managed.id, persisted)
	}

	runner.Run(t, "state", "delete", "--workspace", config.WorkspaceSlug, "--project-id", project.id, managed.id).RequireQuietSuccess(t)
	managed.deleted = true
}

type managedWorkItem struct {
	runner    *Runner
	workspace string
	projectID string
	id        string
	deleted   bool
}

func registerWorkItemCleanup(t *testing.T, runner *Runner, workspace, projectID, id string) *managedWorkItem {
	t.Helper()
	item := &managedWorkItem{runner: runner, workspace: workspace, projectID: projectID, id: id}
	t.Cleanup(func() {
		if item.deleted {
			t.Logf("functional cleanup: work item %s was already deleted by the lifecycle", item.id)
			return
		}
		result := runner.Run(t, "work-item", "delete", "--workspace", workspace, "--project-id", projectID, id)
		if result.Err != nil || result.ExitCode != 0 {
			t.Logf("functional cleanup: generated work item %s was already deleted or could not be deleted: %s", id, result.Output())
		}
	})
	return item
}

func TestPrimaryWorkItemLifecycle(t *testing.T) {
	config, ok := lifecycleConfig(t)
	if !ok {
		return
	}
	runner, err := newRunner(t, config)
	if err != nil {
		t.Fatalf("configure functional runner: %v", err)
	}

	projectName := uniqueName("work-items-project")
	projectIdentifier := uniqueIdentifier()
	projectData, _ := runner.RunJSON(t, "project", "create", "--workspace", config.WorkspaceSlug,
		"--name", projectName, "--identifier", projectIdentifier)
	project := registerProjectCleanup(t, runner, config.WorkspaceSlug, projectID(t, projectData))
	projectIdentifier = projectIdentifierValue(t, projectData)

	firstName := uniqueName("work-item-one")
	firstData, _ := runner.RunJSON(t, "work-item", "create", "--workspace", config.WorkspaceSlug, "--project-id", project.id, "--name", firstName)
	first := registerWorkItemCleanup(t, runner, config.WorkspaceSlug, project.id, workItemID(t, firstData))
	firstSequence := workItemSequence(t, firstData)
	secondName := uniqueName("work-item-two")
	secondData, _ := runner.RunJSON(t, "work-item", "create", "--workspace", config.WorkspaceSlug, "--project-id", project.id, "--name", secondName)
	second := registerWorkItemCleanup(t, runner, config.WorkspaceSlug, project.id, workItemID(t, secondData))
	secondSequence := workItemSequence(t, secondData)
	t.Logf("coverage status: generated project identifier=%s; work-item sequences=%s,%s", projectIdentifier, firstSequence, secondSequence)

	list, _ := runner.RunJSON(t, "work-item", "list", "--workspace", config.WorkspaceSlug, "--project-id", project.id)
	if !jsonContainsString(list, first.id) || !jsonContainsString(list, second.id) {
		t.Fatalf("work-item list omitted generated items %s/%s: %s", first.id, second.id, list)
	}
	aliasList, _ := runner.RunJSON(t, "work-items", "list", "--workspace", config.WorkspaceSlug, "--project-id", project.id)
	if !jsonContainsString(aliasList, first.id) || !jsonContainsString(aliasList, second.id) {
		t.Fatalf("work-items list alias omitted generated items %s/%s: %s", first.id, second.id, aliasList)
	}

	got, _ := runner.RunJSON(t, "work-item", "get", "--workspace", config.WorkspaceSlug, "--project-id", project.id, first.id)
	if !jsonContainsString(got, first.id) || !jsonContainsString(got, firstName) {
		t.Fatalf("work-item get omitted generated item %s: %s", first.id, got)
	}
	identified, _ := runner.RunJSON(t, "work-item", "get-by-identifier", "--workspace", config.WorkspaceSlug, projectIdentifier, firstSequence)
	if !jsonContainsString(identified, first.id) {
		t.Fatalf("identifier lookup omitted generated item %s: %s", first.id, identified)
	}
	searched, _ := runner.RunJSON(t, "work-item", "search", "--workspace", config.WorkspaceSlug, "--search", firstName, "--project-id", project.id)
	if !jsonContainsString(searched, first.id) && !jsonContainsString(searched, firstName) {
		t.Fatalf("work-item search omitted generated item %s/%q: %s", first.id, firstName, searched)
	}

	updatedName := firstName + "-updated"
	updated, _ := runner.RunJSON(t, "work-item", "update", "--workspace", config.WorkspaceSlug, "--project-id", project.id, first.id, "--name", updatedName)
	if !jsonContainsString(updated, updatedName) {
		t.Fatalf("work-item update omitted updated name %q: %s", updatedName, updated)
	}

	relation, _ := runner.RunJSON(t, "work-item", "relations-create", "--workspace", config.WorkspaceSlug, "--project-id", project.id, first.id,
		"--relation-type", "relates_to", "--issue", second.id)
	if len(relation) == 0 {
		t.Fatal("relations-create returned empty JSON")
	}
	relations, _ := runner.RunJSON(t, "work-item", "relations-list", "--workspace", config.WorkspaceSlug, "--project-id", project.id, first.id)
	if len(relations) == 0 {
		t.Fatal("relations-list returned empty JSON")
	}

	runner.Run(t, "work-item", "delete", "--workspace", config.WorkspaceSlug, "--project-id", project.id, second.id).RequireQuietSuccess(t)
	second.deleted = true
	runner.Run(t, "work-item", "delete", "--workspace", config.WorkspaceSlug, "--project-id", project.id, first.id).RequireQuietSuccess(t)
	first.deleted = true
	runner.Run(t, "project", "archive", "--workspace", config.WorkspaceSlug, "--project-id", project.id).RequireQuietSuccess(t)
	project.archived = true
	runner.Run(t, "project", "unarchive", "--workspace", config.WorkspaceSlug, "--project-id", project.id).RequireQuietSuccess(t)
	project.archived = false
	runner.Run(t, "project", "delete", "--workspace", config.WorkspaceSlug, "--project-id", project.id).RequireQuietSuccess(t)
	project.deleted = true
}

func registerLegacyWorkItemCleanup(t *testing.T, runner *Runner, workspace, projectID, id string) *managedWorkItem {
	t.Helper()
	item := &managedWorkItem{runner: runner, workspace: workspace, projectID: projectID, id: id}
	t.Cleanup(func() {
		if item.deleted {
			t.Logf("functional cleanup: legacy work item %s was already deleted by the lifecycle", item.id)
			return
		}
		result := runner.Run(t, "work-item", "legacy", "delete", "--workspace", workspace, "--project-id", projectID, id)
		if result.Err != nil || result.ExitCode != 0 {
			t.Logf("functional cleanup: legacy work item %s was already deleted or could not be deleted: %s", id, result.Output())
		}
	})
	return item
}

func TestLegacyWorkItemCompatibilityWhenOptedIn(t *testing.T) {
	if !isTrue(os.Getenv(functionalOptIn)) || !isTrue(os.Getenv(functionalLegacy)) {
		t.Skipf("coverage status: hidden legacy /issues/ lifecycle skipped; set %s=true and %s=true to enable", functionalOptIn, functionalLegacy)
	}
	config, err := loadFunctionalConfig()
	if err != nil {
		t.Fatalf("legacy functional prerequisites: %v", err)
	}
	runner, err := newRunner(t, config)
	if err != nil {
		t.Fatalf("configure legacy functional runner: %v", err)
	}

	projectData, _ := runner.RunJSON(t, "project", "create", "--workspace", config.WorkspaceSlug,
		"--name", uniqueName("legacy-project"), "--identifier", uniqueIdentifier())
	project := registerProjectCleanup(t, runner, config.WorkspaceSlug, projectID(t, projectData))
	projectIdentifier := projectIdentifierValue(t, projectData)
	itemName := uniqueName("legacy-work-item")
	created, supported := runLegacyJSON(t, runner, "work-item", "legacy", "create", "--workspace", config.WorkspaceSlug, "--project-id", project.id, "--name", itemName)
	if !supported {
		return
	}
	item := registerLegacyWorkItemCleanup(t, runner, config.WorkspaceSlug, project.id, workItemID(t, created))
	sequence := workItemSequence(t, created)

	list, supported := runLegacyJSON(t, runner, "work-item", "legacy", "list", "--workspace", config.WorkspaceSlug, "--project-id", project.id)
	if !supported {
		return
	}
	if !jsonContainsString(list, item.id) {
		t.Fatalf("legacy list omitted generated item %s: %s", item.id, list)
	}
	got, supported := runLegacyJSON(t, runner, "work-item", "legacy", "get", "--workspace", config.WorkspaceSlug, "--project-id", project.id, item.id)
	if !supported {
		return
	}
	if !jsonContainsString(got, item.id) {
		t.Fatalf("legacy get omitted generated item %s: %s", item.id, got)
	}
	identified, supported := runLegacyJSON(t, runner, "work-item", "legacy", "get-by-identifier", "--workspace", config.WorkspaceSlug, projectIdentifier, sequence)
	if !supported {
		return
	}
	if !jsonContainsString(identified, item.id) {
		t.Fatalf("legacy identifier lookup omitted generated item %s: %s", item.id, identified)
	}
	searched, supported := runLegacyJSON(t, runner, "work-item", "legacy", "search", "--workspace", config.WorkspaceSlug,
		"--search", itemName, "--project-id", project.id)
	if !supported {
		return
	}
	if !jsonContainsString(searched, item.id) && !jsonContainsString(searched, itemName) {
		t.Fatalf("legacy search omitted generated item %s/%q: %s", item.id, itemName, searched)
	}
	updatedName := itemName + "-updated"
	updated, supported := runLegacyJSON(t, runner, "work-item", "legacy", "update", "--workspace", config.WorkspaceSlug, "--project-id", project.id, item.id, "--name", updatedName)
	if !supported {
		return
	}
	if !jsonContainsString(updated, updatedName) {
		t.Fatalf("legacy update omitted updated name %q: %s", updatedName, updated)
	}

	runner.Run(t, "work-item", "legacy", "delete", "--workspace", config.WorkspaceSlug, "--project-id", project.id, item.id).RequireQuietSuccess(t)
	item.deleted = true
	runner.Run(t, "project", "archive", "--workspace", config.WorkspaceSlug, "--project-id", project.id).RequireQuietSuccess(t)
	project.archived = true
	runner.Run(t, "project", "unarchive", "--workspace", config.WorkspaceSlug, "--project-id", project.id).RequireQuietSuccess(t)
	project.archived = false
	runner.Run(t, "project", "delete", "--workspace", config.WorkspaceSlug, "--project-id", project.id).RequireQuietSuccess(t)
	project.deleted = true
}

func runLegacyJSON(t *testing.T, runner *Runner, args ...string) (json.RawMessage, bool) {
	t.Helper()
	result := runner.Run(t, append(append([]string(nil), args...), "--output", "json")...)
	if result.Err != nil || result.ExitCode != 0 {
		if legacyRouteUnsupported(result) {
			t.Skipf("coverage status: target does not expose deprecated /issues/ routes; legacy command %s returned an explicit unsupported response", formatArgs(result.Args))
			return nil, false
		}
		result.RequireSuccess(t)
	}
	data := bytes.TrimSpace([]byte(result.Stdout))
	if len(data) == 0 || !json.Valid(data) {
		t.Fatalf("legacy command %s returned invalid JSON: %s", formatArgs(result.Args), result.Output())
	}
	assertNoCredentialMaterial(t, result.Output())
	return append(json.RawMessage(nil), data...), true
}

func legacyRouteUnsupported(result commandResult) bool {
	output := strings.ToLower(result.Output())
	for _, marker := range []string{"404", "405", "not found", "method not allowed", "unsupported", "deprecated route"} {
		if strings.Contains(output, marker) {
			return true
		}
	}
	return false
}

func lifecycleConfig(t *testing.T) (FunctionalConfig, bool) {
	t.Helper()
	if !isTrue(os.Getenv(functionalOptIn)) {
		t.Skipf("functional lifecycle skipped: set %s=true and provide the documented non-production settings", functionalOptIn)
		return FunctionalConfig{}, false
	}
	config, err := loadFunctionalConfig()
	if err != nil {
		t.Fatalf("functional lifecycle prerequisites: %v", err)
	}
	return config, true
}

func jsonContainsString(data json.RawMessage, wanted string) bool {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil {
		return false
	}
	return jsonValueContains(value, wanted)
}

func jsonValueContains(value any, wanted string) bool {
	switch typed := value.(type) {
	case string:
		return typed == wanted
	case json.Number:
		return typed.String() == wanted
	case []any:
		for _, item := range typed {
			if jsonValueContains(item, wanted) {
				return true
			}
		}
	case map[string]any:
		for _, item := range typed {
			if jsonValueContains(item, wanted) {
				return true
			}
		}
	}
	return false
}

func uniqueIdentifier() string {
	return fmt.Sprintf("PF%04X%04X", uint64(os.Getpid())&0xffff, uint64(time.Now().UnixNano())&0xffff)
}

func projectIdentifierValue(t *testing.T, data json.RawMessage) string {
	return jsonStringField(t, data, "identifier")
}

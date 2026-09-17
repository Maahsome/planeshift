package project

import (
	projectresource "planeshift/projects"

	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "update",
		Args: cobra.NoArgs,
		RunE: runUpdate,
	}
	addContextFlags(command, true)
	addUpdateFlags(command)
	return command
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name, err := optionalStringFlag(cmd, "name")
	if err != nil {
		return err
	}
	description, err := optionalStringFlag(cmd, "description")
	if err != nil {
		return err
	}
	projectLead, err := optionalStringFlag(cmd, "project-lead")
	if err != nil {
		return err
	}
	defaultAssignee, err := optionalStringFlag(cmd, "default-assignee")
	if err != nil {
		return err
	}
	identifier, err := optionalStringFlag(cmd, "identifier")
	if err != nil {
		return err
	}
	iconProp, err := optionalJSONFlag(cmd, "icon-prop")
	if err != nil {
		return err
	}
	emoji, err := optionalStringFlag(cmd, "emoji")
	if err != nil {
		return err
	}
	coverImage, err := optionalStringFlag(cmd, "cover-image")
	if err != nil {
		return err
	}
	moduleView, err := optionalBoolFlag(cmd, "module-view")
	if err != nil {
		return err
	}
	cycleView, err := optionalBoolFlag(cmd, "cycle-view")
	if err != nil {
		return err
	}
	issueViewsView, err := optionalBoolFlag(cmd, "issue-views-view")
	if err != nil {
		return err
	}
	pageView, err := optionalBoolFlag(cmd, "page-view")
	if err != nil {
		return err
	}
	intakeView, err := optionalBoolFlag(cmd, "intake-view")
	if err != nil {
		return err
	}
	guestViewAllFeatures, err := optionalBoolFlag(cmd, "guest-view-all-features")
	if err != nil {
		return err
	}
	archiveIn, err := optionalIntFlag(cmd, "archive-in")
	if err != nil {
		return err
	}
	closeIn, err := optionalIntFlag(cmd, "close-in")
	if err != nil {
		return err
	}
	timezone, err := optionalStringFlag(cmd, "timezone")
	if err != nil {
		return err
	}
	externalSource, err := optionalStringFlag(cmd, "external-source")
	if err != nil {
		return err
	}
	externalID, err := optionalStringFlag(cmd, "external-id")
	if err != nil {
		return err
	}
	isIssueTypeEnabled, err := optionalBoolFlag(cmd, "is-issue-type-enabled")
	if err != nil {
		return err
	}
	isTimeTrackingEnabled, err := optionalBoolFlag(cmd, "is-time-tracking-enabled")
	if err != nil {
		return err
	}
	defaultState, err := optionalStringFlag(cmd, "default-state")
	if err != nil {
		return err
	}
	estimate, err := optionalStringFlag(cmd, "estimate")
	if err != nil {
		return err
	}
	request := projectresource.UpdateProjectRequest{
		Name: name, Description: description, ProjectLead: projectLead, DefaultAssignee: defaultAssignee,
		Identifier: identifier, IconProp: iconProp, Emoji: emoji, CoverImage: coverImage,
		ModuleView: moduleView, CycleView: cycleView, IssueViewsView: issueViewsView,
		PageView: pageView, IntakeView: intakeView, GuestViewAllFeatures: guestViewAllFeatures,
		ArchiveIn: archiveIn, CloseIn: closeIn, Timezone: timezone, ExternalSource: externalSource,
		ExternalID: externalID, IsIssueTypeEnabled: isIssueTypeEnabled,
		IsTimeTrackingEnabled: isTimeTrackingEnabled, DefaultState: defaultState, Estimate: estimate,
	}
	route, err := routeContext(cmd, true)
	if err != nil {
		return err
	}
	client, err := projectClient()
	if err != nil {
		return err
	}
	project, _, err := client.Update(cmd.Context(), route.Workspace, route.ProjectID, request)
	if err != nil {
		return err
	}
	return outputProject(project)
}

func addUpdateFlags(command *cobra.Command) {
	command.Flags().String("name", "", "Project name")
	command.Flags().String("description", "", "Project description")
	command.Flags().String("project-lead", "", "Project lead user ID")
	command.Flags().String("default-assignee", "", "Default assignee user ID")
	command.Flags().String("identifier", "", "Project identifier")
	command.Flags().String("icon-prop", "", "Project icon JSON")
	command.Flags().String("emoji", "", "Project emoji")
	command.Flags().String("cover-image", "", "Project cover image URL")
	command.Flags().Bool("module-view", false, "Enable module view")
	command.Flags().Bool("cycle-view", false, "Enable cycle view")
	command.Flags().Bool("issue-views-view", false, "Enable project views")
	command.Flags().Bool("page-view", false, "Enable page view")
	command.Flags().Bool("intake-view", false, "Enable intake view")
	command.Flags().Bool("guest-view-all-features", false, "Enable guest access to all features")
	command.Flags().Int("archive-in", 0, "Months before automatic archive")
	command.Flags().Int("close-in", 0, "Months before automatic close")
	command.Flags().String("timezone", "", "Project timezone")
	command.Flags().String("external-source", "", "External source")
	command.Flags().String("external-id", "", "External ID")
	command.Flags().Bool("is-issue-type-enabled", false, "Enable issue types")
	command.Flags().Bool("is-time-tracking-enabled", false, "Enable time tracking")
	command.Flags().String("default-state", "", "Default state ID")
	command.Flags().String("estimate", "", "Estimate ID")
}

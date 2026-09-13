package get

import (
	"encoding/json"
	"fmt"

	"planeshift/help"
	projectresource "planeshift/projects"

	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: (&help.ProjectsCmd{}).Short(),
	Long:  (&help.ProjectsCmd{}).Long(),
	Args:  cobra.NoArgs,
}

func projectClient() (*projectresource.Client, error) {
	client, err := clientFactory.New()
	if err != nil {
		return nil, err
	}
	return projectresource.NewClient(client), nil
}

func outputProject(value any) error {
	if c == nil {
		return fmt.Errorf("project command configuration is not initialized")
	}
	output, err := projectresource.RawOutput(value)
	if err != nil {
		return err
	}
	if c.OutputFormat == "" {
		c.OutputFormat = "json"
	}
	c.OutputData(output)
	return nil
}

func optionalStringFlag(cmd *cobra.Command, name string) (*string, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func optionalBoolFlag(cmd *cobra.Command, name string) (*bool, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetBool(name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func optionalIntFlag(cmd *cobra.Command, name string) (*int, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetInt(name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func optionalJSONFlag(cmd *cobra.Command, name string) (*json.RawMessage, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return nil, err
	}
	data := json.RawMessage(value)
	if !json.Valid(data) {
		return nil, fmt.Errorf("--%s must contain valid JSON", name)
	}
	return &data, nil
}

func runList(cmd *cobra.Command, args []string) error {
	cursor, err := cmd.Flags().GetString("cursor")
	if err != nil {
		return err
	}
	perPage, err := cmd.Flags().GetInt("per-page")
	if err != nil {
		return err
	}
	fields, err := cmd.Flags().GetString("fields")
	if err != nil {
		return err
	}
	expand, err := cmd.Flags().GetString("expand")
	if err != nil {
		return err
	}
	orderBy, err := cmd.Flags().GetString("order-by")
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("per-page") && perPage == 0 {
		return fmt.Errorf("--per-page must be between 1 and 100")
	}
	options := projectresource.ListOptions{
		Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand, OrderBy: orderBy,
	}
	if _, err := options.Query(); err != nil {
		return err
	}
	client, err := projectClient()
	if err != nil {
		return err
	}
	page, _, err := client.List(cmd.Context(), args[0], options)
	if err != nil {
		return err
	}
	return outputProject(page)
}

func runCreate(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	identifier, err := cmd.Flags().GetString("identifier")
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
	request := projectresource.CreateProjectRequest{
		Name: name, Identifier: identifier, Description: description, ProjectLead: projectLead,
		DefaultAssignee: defaultAssignee, IconProp: iconProp, Emoji: emoji, CoverImage: coverImage,
		ModuleView: moduleView, CycleView: cycleView, IssueViewsView: issueViewsView,
		PageView: pageView, IntakeView: intakeView, GuestViewAllFeatures: guestViewAllFeatures,
		ArchiveIn: archiveIn, CloseIn: closeIn, Timezone: timezone, ExternalSource: externalSource,
		ExternalID: externalID, IsIssueTypeEnabled: isIssueTypeEnabled,
		IsTimeTrackingEnabled: isTimeTrackingEnabled,
	}
	client, err := projectClient()
	if err != nil {
		return err
	}
	project, _, err := client.Create(cmd.Context(), args[0], request)
	if err != nil {
		return err
	}
	return outputProject(project)
}

func runCreateTemplate(cmd *cobra.Command, args []string) error {
	templateID, err := cmd.Flags().GetString("template-id")
	if err != nil {
		return err
	}
	name, err := optionalStringFlag(cmd, "name")
	if err != nil {
		return err
	}
	identifier, err := optionalStringFlag(cmd, "identifier")
	if err != nil {
		return err
	}
	description, err := optionalStringFlag(cmd, "description")
	if err != nil {
		return err
	}
	network, err := optionalIntFlag(cmd, "network")
	if err != nil {
		return err
	}
	projectLead, err := optionalStringFlag(cmd, "project-lead")
	if err != nil {
		return err
	}
	request := projectresource.CreateProjectFromTemplateRequest{
		TemplateID: templateID, Name: name, Identifier: identifier, Description: description,
		Network: network, ProjectLead: projectLead,
	}
	client, err := projectClient()
	if err != nil {
		return err
	}
	project, _, err := client.CreateFromTemplate(cmd.Context(), args[0], request)
	if err != nil {
		return err
	}
	return outputProject(project)
}

func runGet(cmd *cobra.Command, args []string) error {
	client, err := projectClient()
	if err != nil {
		return err
	}
	project, _, err := client.Get(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}
	return outputProject(project)
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
	client, err := projectClient()
	if err != nil {
		return err
	}
	project, _, err := client.Update(cmd.Context(), args[0], args[1], request)
	if err != nil {
		return err
	}
	return outputProject(project)
}

func runArchive(cmd *cobra.Command, args []string) error {
	client, err := projectClient()
	if err != nil {
		return err
	}
	_, err = client.Archive(cmd.Context(), args[0], args[1])
	return err
}

func runUnarchive(cmd *cobra.Command, args []string) error {
	client, err := projectClient()
	if err != nil {
		return err
	}
	_, err = client.Unarchive(cmd.Context(), args[0], args[1])
	return err
}

func runDelete(cmd *cobra.Command, args []string) error {
	client, err := projectClient()
	if err != nil {
		return err
	}
	_, err = client.Delete(cmd.Context(), args[0], args[1])
	return err
}

func addListFlags(command *cobra.Command) {
	command.Flags().String("cursor", "", "Cursor for the next or previous project page")
	command.Flags().Int("per-page", 0, "Projects per page (Plane default 20; valid range 1-100)")
	command.Flags().String("fields", "", "Comma-separated project fields to return")
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
	command.Flags().String("order-by", "", "Project ordering field; prefix with - for descending")
}

func addCreateFlags(command *cobra.Command) {
	command.Flags().String("name", "", "Project name")
	command.Flags().String("identifier", "", "Project identifier")
	command.Flags().String("description", "", "Project description")
	command.Flags().String("project-lead", "", "Project lead user ID")
	command.Flags().String("default-assignee", "", "Default assignee user ID")
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

func init() {
	list := &cobra.Command{Use: "list workspace_slug", Args: cobra.ExactArgs(1), RunE: runList}
	addListFlags(list)

	create := &cobra.Command{Use: "create workspace_slug", Args: cobra.ExactArgs(1), RunE: runCreate}
	addCreateFlags(create)
	_ = create.MarkFlagRequired("name")
	_ = create.MarkFlagRequired("identifier")

	createTemplate := &cobra.Command{Use: "create-template workspace_slug", Args: cobra.ExactArgs(1), RunE: runCreateTemplate}
	createTemplate.Flags().String("template-id", "", "Project template ID")
	createTemplate.Flags().String("name", "", "Project name override")
	createTemplate.Flags().String("identifier", "", "Project identifier override")
	createTemplate.Flags().String("description", "", "Project description override")
	createTemplate.Flags().Int("network", 0, "Project visibility: 0 secret or 2 public")
	createTemplate.Flags().String("project-lead", "", "Project lead user ID")
	_ = createTemplate.MarkFlagRequired("template-id")

	get := &cobra.Command{Use: "get workspace_slug project_id", Args: cobra.ExactArgs(2), RunE: runGet}
	update := &cobra.Command{Use: "update workspace_slug project_id", Args: cobra.ExactArgs(2), RunE: runUpdate}
	addUpdateFlags(update)
	archive := &cobra.Command{Use: "archive workspace_slug project_id", Args: cobra.ExactArgs(2), RunE: runArchive}
	unarchive := &cobra.Command{Use: "unarchive workspace_slug project_id", Args: cobra.ExactArgs(2), RunE: runUnarchive}
	deleteProject := &cobra.Command{Use: "delete workspace_slug project_id", Args: cobra.ExactArgs(2), RunE: runDelete}

	projectsCmd.AddCommand(list, create, createTemplate, get, update, archive, unarchive, deleteProject)
	getCmd.AddCommand(projectsCmd)
}

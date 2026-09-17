package project

import (
	projectresource "planeshift/projects"

	"github.com/spf13/cobra"
)

func newCreateTemplateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "create-template",
		Args: cobra.NoArgs,
		RunE: runCreateTemplate,
	}
	addContextFlags(command, false)
	command.Flags().String("template-id", "", "Project template ID")
	command.Flags().String("name", "", "Project name override")
	command.Flags().String("identifier", "", "Project identifier override")
	command.Flags().String("description", "", "Project description override")
	command.Flags().Int("network", 0, "Project visibility: 0 secret or 2 public")
	command.Flags().String("project-lead", "", "Project lead user ID")
	_ = command.MarkFlagRequired("template-id")
	return command
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
	route, err := routeContext(cmd, false)
	if err != nil {
		return err
	}
	client, err := projectClient()
	if err != nil {
		return err
	}
	project, _, err := client.CreateFromTemplate(cmd.Context(), route.Workspace, request)
	if err != nil {
		return err
	}
	return outputProject(project)
}

package context

import (
	"fmt"
	"strings"

	"planeshift/config"
	"planeshift/help"
	"planeshift/plane"
	projectresource "planeshift/projects"

	"github.com/spf13/cobra"
)

func newSetCommand(conf *config.Config, factory plane.ClientFactory, saver Saver, prompt Prompter) *cobra.Command {
	command := &cobra.Command{
		Use:   "set",
		Short: (&help.ContextSetCmd{}).Short(),
		Long:  (&help.ContextSetCmd{}).Long(),
		Args:  cobra.NoArgs,
	}
	command.Flags().String("workspace", "", "Workspace slug to store in context")
	command.Flags().String("project", "", "Project ID to store in context")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		return runSet(cmd, conf, factory, saver, prompt)
	}
	return command
}

func runSet(cmd *cobra.Command, conf *config.Config, factory plane.ClientFactory, saver Saver, prompt Prompter) error {
	if conf == nil {
		return fmt.Errorf("context command configuration is not initialized")
	}
	if saver == nil {
		return fmt.Errorf("context saver is not initialized")
	}

	workspaceChanged := cmd.Flags().Changed("workspace")
	projectChanged := cmd.Flags().Changed("project")
	candidate := conf.Context
	if workspaceChanged {
		workspace, err := cmd.Flags().GetString("workspace")
		if err != nil {
			return err
		}
		candidate.Workspace, err = requiredContextValue("workspace", workspace)
		if err != nil {
			return err
		}
	}
	if projectChanged {
		projectID, err := cmd.Flags().GetString("project")
		if err != nil {
			return err
		}
		candidate.Project.ID, err = requiredContextValue("project", projectID)
		if err != nil {
			return err
		}
	}
	if workspaceChanged || projectChanged {
		return saver(candidate)
	}

	return runInteractiveSet(cmd, conf, factory, saver, prompt)
}

func runInteractiveSet(cmd *cobra.Command, conf *config.Config, factory plane.ClientFactory, saver Saver, prompt Prompter) error {
	workspace, err := prompt.Workspace()
	if err != nil {
		return fmt.Errorf("prompt for workspace: %w", err)
	}
	workspace, err = requiredContextValue("workspace", workspace)
	if err != nil {
		return err
	}

	client, err := factory.New()
	if err != nil {
		return err
	}
	projectsClient := projectresource.NewClient(client)
	page, _, err := projectsClient.List(cmd.Context(), workspace, projectresource.ListOptions{})
	if err != nil {
		return err
	}
	if len(page.Results) == 0 {
		return fmt.Errorf("no projects found in workspace %q", workspace)
	}

	names := make([]string, 0, len(page.Results))
	for _, project := range page.Results {
		names = append(names, project.Name)
	}
	selected, err := prompt.Project(names)
	if err != nil {
		return fmt.Errorf("prompt for project: %w", err)
	}
	project, found := findProject(page.Results, selected)
	if !found {
		return fmt.Errorf("selected project %q was not returned by Plane", selected)
	}
	if strings.TrimSpace(project.ID) == "" || strings.TrimSpace(project.Name) == "" {
		return fmt.Errorf("selected project has incomplete identity")
	}

	return saver(config.Context{
		Workspace: workspace,
		Project: config.ProjectContext{
			ID:   strings.TrimSpace(project.ID),
			Name: strings.TrimSpace(project.Name),
		},
	})
}

func requiredContextValue(name, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("--%s must not be blank", name)
	}
	return value, nil
}

func findProject(values []projectresource.Project, name string) (projectresource.Project, bool) {
	for _, project := range values {
		if project.Name == name {
			return project, true
		}
	}
	return projectresource.Project{}, false
}

package project

import (
	projectfeatures "planeshift/projectfeatures"

	"github.com/spf13/cobra"
)

func newFeaturesUpdateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "update workspace_slug project_id",
		Args: cobra.ExactArgs(2),
		RunE: runFeaturesUpdate,
	}
	addFeaturesUpdateFlags(command)
	return command
}

func runFeaturesUpdate(cmd *cobra.Command, args []string) error {
	epics, err := optionalBoolFlag(cmd, "epics")
	if err != nil {
		return err
	}
	modules, err := optionalBoolFlag(cmd, "modules")
	if err != nil {
		return err
	}
	cycles, err := optionalBoolFlag(cmd, "cycles")
	if err != nil {
		return err
	}
	views, err := optionalBoolFlag(cmd, "views")
	if err != nil {
		return err
	}
	pages, err := optionalBoolFlag(cmd, "pages")
	if err != nil {
		return err
	}
	intakes, err := optionalBoolFlag(cmd, "intakes")
	if err != nil {
		return err
	}
	workItemTypes, err := optionalBoolFlag(cmd, "work-item-types")
	if err != nil {
		return err
	}

	request := projectfeatures.UpdateProjectFeaturesRequest{
		Epics: epics, Modules: modules, Cycles: cycles, Views: views,
		Pages: pages, Intakes: intakes, WorkItemTypes: workItemTypes,
	}
	client, err := projectFeaturesClient()
	if err != nil {
		return err
	}
	features, response, err := client.Update(cmd.Context(), args[0], args[1], request)
	if err != nil {
		return err
	}
	if response.StatusCode == 204 {
		return nil
	}
	return outputProjectFeatures(features)
}

func addFeaturesUpdateFlags(command *cobra.Command) {
	command.Flags().Bool("epics", false, "Enable or disable epics")
	command.Flags().Bool("modules", false, "Enable or disable modules")
	command.Flags().Bool("cycles", false, "Enable or disable cycles")
	command.Flags().Bool("views", false, "Enable or disable views")
	command.Flags().Bool("pages", false, "Enable or disable pages")
	command.Flags().Bool("intakes", false, "Enable or disable intakes")
	command.Flags().Bool("work-item-types", false, "Enable or disable work item types")
}

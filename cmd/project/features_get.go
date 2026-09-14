package project

import "github.com/spf13/cobra"

func newFeaturesGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "get workspace_slug project_id",
		Args: cobra.ExactArgs(2),
		RunE: runFeaturesGet,
	}
}

func runFeaturesGet(cmd *cobra.Command, args []string) error {
	client, err := projectFeaturesClient()
	if err != nil {
		return err
	}
	features, response, err := client.Get(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}
	if response.StatusCode == 204 {
		return nil
	}
	return outputProjectFeatures(features)
}

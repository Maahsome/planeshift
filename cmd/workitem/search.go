package workitem

import "github.com/spf13/cobra"

func newSearchCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "search workspace_slug", Args: cobra.ExactArgs(1), RunE: runSearch}
	addSearchFlags(command)
	if legacy {
		command.RunE = runLegacySearch
	}
	return command
}

func runSearch(cmd *cobra.Command, args []string) error {
	options, err := searchOptions(cmd)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	result, _, err := client.Search(cmd.Context(), args[0], options)
	if err != nil {
		return err
	}
	return outputWorkItem(result)
}

func runLegacySearch(cmd *cobra.Command, args []string) error {
	options, err := searchOptions(cmd)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	result, _, err := client.LegacySearch(cmd.Context(), args[0], options)
	if err != nil {
		return err
	}
	return outputWorkItem(result)
}

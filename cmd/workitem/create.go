package workitem

import "github.com/spf13/cobra"

func newCreateCommand(legacy bool) *cobra.Command {
	command := &cobra.Command{Use: "create workspace_slug project_id", Args: cobra.ExactArgs(2), RunE: runCreate}
	addRequestFlags(command, false)
	if legacy {
		command.RunE = runLegacyCreate
	}
	return command
}

func runCreate(cmd *cobra.Command, args []string) error {
	request, err := requestValues(cmd)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	item, _, err := client.Create(cmd.Context(), args[0], args[1], request)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}

func runLegacyCreate(cmd *cobra.Command, args []string) error {
	request, err := requestValues(cmd)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	item, _, err := client.LegacyCreate(cmd.Context(), args[0], args[1], request)
	if err != nil {
		return err
	}
	return outputWorkItem(item)
}

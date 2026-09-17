package workitem

import (
	"fmt"
	"strings"

	resource "planeshift/workitems"

	"github.com/spf13/cobra"
)

func newRelationsCreateCommand() *cobra.Command {
	command := &cobra.Command{Use: "relations-create work_item_id", Args: cobra.ExactArgs(1), RunE: runRelationsCreate}
	addContextFlags(command)
	command.Flags().String("relation-type", "", "Relation type")
	command.Flags().StringSlice("issue", nil, "Related work-item ID; repeat for multiple IDs")
	_ = command.MarkFlagRequired("relation-type")
	_ = command.MarkFlagRequired("issue")
	return command
}

func runRelationsCreate(cmd *cobra.Command, args []string) error {
	relationType, err := cmd.Flags().GetString("relation-type")
	if err != nil {
		return err
	}
	issues, err := cmd.Flags().GetStringSlice("issue")
	if err != nil {
		return err
	}
	if len(issues) == 0 || (len(issues) == 1 && strings.TrimSpace(issues[0]) == "") {
		return fmt.Errorf("--issue is required")
	}
	request := resource.CreateWorkItemRelationRequest{RelationType: resource.RelationType(relationType), Issues: issues}
	if err := validateRelationRequest(request); err != nil {
		return err
	}
	route, err := routeContext(cmd, true)
	if err != nil {
		return err
	}
	client, err := workItemClient()
	if err != nil {
		return err
	}
	result, _, err := client.CreateRelation(cmd.Context(), route.Workspace, route.ProjectID, args[0], request)
	if err != nil {
		return err
	}
	return outputWorkItem(result)
}

func validateRelationRequest(request resource.CreateWorkItemRelationRequest) error {
	data := map[resource.RelationType]bool{
		resource.RelationBlocking: true, resource.RelationBlockedBy: true,
		resource.RelationDuplicate: true, resource.RelationRelatesTo: true,
		resource.RelationStartBefore: true, resource.RelationStartAfter: true,
		resource.RelationFinishBefore: true, resource.RelationFinishAfter: true,
	}
	if !data[request.RelationType] {
		return fmt.Errorf("relation_type %q is not documented", request.RelationType)
	}
	if len(request.Issues) == 0 {
		return fmt.Errorf("relation issues are required")
	}
	return nil
}

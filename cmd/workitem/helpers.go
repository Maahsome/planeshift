package workitem

import (
	"encoding/json"
	"fmt"

	"planeshift/objects"
	resource "planeshift/workitems"

	"github.com/spf13/cobra"
)

func workItemClient() (*resource.Client, error) {
	if clientFactory == nil {
		return nil, fmt.Errorf("work item client factory is not initialized")
	}
	client, err := clientFactory.New()
	if err != nil {
		return nil, err
	}
	return resource.NewClient(client), nil
}

func outputWorkItem(value any) error {
	if c == nil {
		return fmt.Errorf("work item command configuration is not initialized")
	}
	output, err := objects.NewRawJSONFromValue(value)
	if err != nil {
		return err
	}
	if c.OutputFormat == "" {
		c.OutputFormat = "json"
	}
	c.OutputData(output)
	return nil
}

func optionalString(cmd *cobra.Command, name string) (*string, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetString(name)
	return &value, err
}

func optionalBool(cmd *cobra.Command, name string) (*bool, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetBool(name)
	return &value, err
}

func optionalInt(cmd *cobra.Command, name string) (*int, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetInt(name)
	return &value, err
}

func optionalStrings(cmd *cobra.Command, name string) (*[]string, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetStringSlice(name)
	return &value, err
}

func optionalJSON(cmd *cobra.Command, name string) (*json.RawMessage, error) {
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

// optionalDynamicJSON accepts documented scalar strings as well as raw JSON
// for fields whose API response can be expanded or null.
func optionalDynamicJSON(cmd *cobra.Command, name string) (*json.RawMessage, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return nil, err
	}
	data := json.RawMessage(value)
	if !json.Valid(data) {
		encoded, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			return nil, fmt.Errorf("encode --%s", name)
		}
		data = encoded
	}
	return &data, nil
}

func addListFlags(command *cobra.Command) {
	command.Flags().String("cursor", "", "Cursor for the next or previous work-item page")
	command.Flags().Int("per-page", 0, "Work items per page (Plane default 20; valid range 1-100)")
	command.Flags().String("fields", "", "Comma-separated work-item fields to return")
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
	command.Flags().String("external-id", "", "External system identifier")
	command.Flags().String("external-source", "", "External system source")
	command.Flags().String("order-by", "", "Work-item ordering field; prefix with - for descending")
}

func listOptions(cmd *cobra.Command) (resource.ListOptions, error) {
	perPage, err := cmd.Flags().GetInt("per-page")
	if err != nil {
		return resource.ListOptions{}, err
	}
	if cmd.Flags().Changed("per-page") && perPage == 0 {
		return resource.ListOptions{}, fmt.Errorf("--per-page must be between 1 and 100")
	}
	cursor, err := cmd.Flags().GetString("cursor")
	if err != nil {
		return resource.ListOptions{}, err
	}
	fields, err := cmd.Flags().GetString("fields")
	if err != nil {
		return resource.ListOptions{}, err
	}
	expand, err := cmd.Flags().GetString("expand")
	if err != nil {
		return resource.ListOptions{}, err
	}
	externalID, err := cmd.Flags().GetString("external-id")
	if err != nil {
		return resource.ListOptions{}, err
	}
	externalSource, err := cmd.Flags().GetString("external-source")
	if err != nil {
		return resource.ListOptions{}, err
	}
	orderBy, err := cmd.Flags().GetString("order-by")
	if err != nil {
		return resource.ListOptions{}, err
	}
	options := resource.ListOptions{Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand,
		ExternalID: externalID, ExternalSource: externalSource, OrderBy: orderBy}
	_, err = options.Query()
	return options, err
}

func addDetailFlags(command *cobra.Command) {
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
	command.Flags().String("fields", "", "Comma-separated fields to return")
	command.Flags().String("external-id", "", "External system identifier")
	command.Flags().String("external-source", "", "External system source")
	command.Flags().String("order-by", "", "Ordering field; prefix with - for descending")
}

func detailOptions(cmd *cobra.Command) (resource.DetailOptions, error) {
	expand, err := cmd.Flags().GetString("expand")
	if err != nil {
		return resource.DetailOptions{}, err
	}
	fields, err := cmd.Flags().GetString("fields")
	if err != nil {
		return resource.DetailOptions{}, err
	}
	externalID, err := cmd.Flags().GetString("external-id")
	if err != nil {
		return resource.DetailOptions{}, err
	}
	externalSource, err := cmd.Flags().GetString("external-source")
	if err != nil {
		return resource.DetailOptions{}, err
	}
	orderBy, err := cmd.Flags().GetString("order-by")
	if err != nil {
		return resource.DetailOptions{}, err
	}
	return resource.DetailOptions{Expand: expand, Fields: fields, ExternalID: externalID,
		ExternalSource: externalSource, OrderBy: orderBy}, nil
}

func addSearchFlags(command *cobra.Command) {
	command.Flags().String("search", "", "Required text to search for")
	command.Flags().Int("limit", 0, "Maximum search results")
	command.Flags().String("project-id", "", "Limit search to a project")
	command.Flags().String("workspace-search", "", "Whether to search across the workspace")
	_ = command.MarkFlagRequired("search")
}

func searchOptions(cmd *cobra.Command) (resource.SearchOptions, error) {
	search, err := cmd.Flags().GetString("search")
	if err != nil {
		return resource.SearchOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return resource.SearchOptions{}, err
	}
	projectID, err := cmd.Flags().GetString("project-id")
	if err != nil {
		return resource.SearchOptions{}, err
	}
	workspaceSearch, err := cmd.Flags().GetString("workspace-search")
	if err != nil {
		return resource.SearchOptions{}, err
	}
	options := resource.SearchOptions{Search: search, Limit: limit, ProjectID: projectID, WorkspaceSearch: workspaceSearch}
	_, err = options.Query()
	return options, err
}

func addRelationListFlags(command *cobra.Command) {
	command.Flags().String("cursor", "", "Cursor for the next or previous relation page")
	command.Flags().Int("per-page", 0, "Relations per page (valid range 1-100)")
	command.Flags().String("fields", "", "Comma-separated relation fields to return")
	command.Flags().String("expand", "", "Comma-separated related fields to expand")
}

func relationListOptions(cmd *cobra.Command) (resource.RelationListOptions, error) {
	perPage, err := cmd.Flags().GetInt("per-page")
	if err != nil {
		return resource.RelationListOptions{}, err
	}
	if cmd.Flags().Changed("per-page") && perPage == 0 {
		return resource.RelationListOptions{}, fmt.Errorf("--per-page must be between 1 and 100")
	}
	cursor, err := cmd.Flags().GetString("cursor")
	if err != nil {
		return resource.RelationListOptions{}, err
	}
	fields, err := cmd.Flags().GetString("fields")
	if err != nil {
		return resource.RelationListOptions{}, err
	}
	expand, err := cmd.Flags().GetString("expand")
	if err != nil {
		return resource.RelationListOptions{}, err
	}
	options := resource.RelationListOptions{Cursor: cursor, PerPage: perPage, Fields: fields, Expand: expand}
	_, err = options.Query()
	return options, err
}

func addRequestFlags(command *cobra.Command, update bool) {
	command.Flags().StringSlice("assignees", nil, "Assignee user IDs")
	command.Flags().StringSlice("labels", nil, "Label IDs")
	command.Flags().String("type-id", "", "Work-item type ID")
	command.Flags().String("parent", "", "Parent work-item ID")
	command.Flags().String("deleted-at", "", "Deleted-at timestamp")
	command.Flags().Int("point", 0, "Point value")
	if update {
		command.Flags().String("name", "", "Work-item name")
	}
	command.Flags().String("description-html", "", "HTML description")
	command.Flags().String("description-stripped", "", "Stripped description")
	command.Flags().String("priority", "", "Priority: urgent, high, medium, low, or none")
	command.Flags().String("start-date", "", "Start date")
	command.Flags().String("target-date", "", "Target date")
	command.Flags().Int("sequence-id", 0, "Sequence ID")
	command.Flags().String("sort-order", "", "Sort-order JSON value")
	command.Flags().String("completed-at", "", "Completion timestamp")
	command.Flags().String("archived-at", "", "Archive timestamp")
	command.Flags().String("last-activity-at", "", "Last activity timestamp")
	command.Flags().Bool("is-draft", false, "Whether the work item is a draft")
	command.Flags().String("external-source", "", "External source")
	command.Flags().String("external-id", "", "External ID")
	command.Flags().String("created-by", "", "Creator ID")
	command.Flags().String("state", "", "State ID")
	command.Flags().String("estimate-point", "", "Estimate-point JSON value")
	command.Flags().String("type", "", "Work-item type JSON value")
	if !update {
		command.Flags().String("name", "", "Work-item name")
		_ = command.MarkFlagRequired("name")
	}
}

func requestValues(cmd *cobra.Command) (resource.CreateWorkItemRequest, error) {
	assignees, err := optionalStrings(cmd, "assignees")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	labels, err := optionalStrings(cmd, "labels")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	typeID, err := optionalString(cmd, "type-id")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	parent, err := optionalString(cmd, "parent")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	deletedAt, err := optionalString(cmd, "deleted-at")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	point, err := optionalInt(cmd, "point")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	descriptionHTML, err := optionalString(cmd, "description-html")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	descriptionStripped, err := optionalString(cmd, "description-stripped")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	priority, err := optionalString(cmd, "priority")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	startDate, err := optionalString(cmd, "start-date")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	targetDate, err := optionalString(cmd, "target-date")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	sequenceID, err := optionalInt(cmd, "sequence-id")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	sortOrder, err := optionalDynamicJSON(cmd, "sort-order")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	completedAt, err := optionalString(cmd, "completed-at")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	archivedAt, err := optionalString(cmd, "archived-at")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	lastActivityAt, err := optionalString(cmd, "last-activity-at")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	isDraft, err := optionalBool(cmd, "is-draft")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	externalSource, err := optionalString(cmd, "external-source")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	externalID, err := optionalString(cmd, "external-id")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	createdBy, err := optionalString(cmd, "created-by")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	state, err := optionalString(cmd, "state")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	estimatePoint, err := optionalDynamicJSON(cmd, "estimate-point")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	typeValue, err := optionalDynamicJSON(cmd, "type")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return resource.CreateWorkItemRequest{}, err
	}
	return resource.CreateWorkItemRequest{
		Assignees: assignees, Labels: labels, TypeID: typeID, Parent: parent, DeletedAt: deletedAt,
		Point: point, Name: name, DescriptionHTML: descriptionHTML, DescriptionStripped: descriptionStripped,
		Priority: priority, StartDate: startDate, TargetDate: targetDate, SequenceID: sequenceID,
		SortOrder: sortOrder, CompletedAt: completedAt, ArchivedAt: archivedAt, LastActivityAt: lastActivityAt,
		IsDraft: isDraft, ExternalSource: externalSource, ExternalID: externalID, CreatedBy: createdBy,
		State: state, EstimatePoint: estimatePoint, Type: typeValue,
	}, nil
}

func updateRequestValues(cmd *cobra.Command) (resource.UpdateWorkItemRequest, error) {
	createRequest, err := requestValues(cmd)
	if err != nil {
		return resource.UpdateWorkItemRequest{}, err
	}
	name, err := optionalString(cmd, "name")
	if err != nil {
		return resource.UpdateWorkItemRequest{}, err
	}
	return resource.UpdateWorkItemRequest{
		Assignees: createRequest.Assignees, Labels: createRequest.Labels, TypeID: createRequest.TypeID,
		Parent: createRequest.Parent, DeletedAt: createRequest.DeletedAt, Point: createRequest.Point,
		Name: name, DescriptionHTML: createRequest.DescriptionHTML, DescriptionStripped: createRequest.DescriptionStripped,
		Priority: createRequest.Priority, StartDate: createRequest.StartDate, TargetDate: createRequest.TargetDate,
		SequenceID: createRequest.SequenceID, SortOrder: createRequest.SortOrder, CompletedAt: createRequest.CompletedAt,
		ArchivedAt: createRequest.ArchivedAt, LastActivityAt: createRequest.LastActivityAt, IsDraft: createRequest.IsDraft,
		ExternalSource: createRequest.ExternalSource, ExternalID: createRequest.ExternalID, CreatedBy: createRequest.CreatedBy,
		State: createRequest.State, EstimatePoint: createRequest.EstimatePoint, Type: createRequest.Type,
	}, nil
}

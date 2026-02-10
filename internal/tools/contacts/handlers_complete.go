package contacts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/people/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- batch_create_contacts (complete) ---

type BatchCreateContactsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Contacts  string `json:"contacts" jsonschema:"required" jsonschema_description:"JSON array of contact objects. Each with given_name family_name email phone org_name org_title."`
}

type ContactEntry struct {
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	OrgName    string `json:"org_name"`
	OrgTitle   string `json:"org_title"`
}

func createBatchCreateContactsHandler(factory *services.Factory) mcp.ToolHandlerFor[BatchCreateContactsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchCreateContactsInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		var entries []ContactEntry
		if err := json.Unmarshal([]byte(input.Contacts), &entries); err != nil {
			return nil, nil, fmt.Errorf("invalid contacts JSON - provide array of {given_name, family_name, email, phone, org_name, org_title}: %w", err)
		}

		if len(entries) > 200 {
			return nil, nil, fmt.Errorf("maximum 200 contacts per batch, got %d", len(entries))
		}

		batchReq := &people.BatchCreateContactsRequest{}
		for _, e := range entries {
			batchReq.Contacts = append(batchReq.Contacts, &people.ContactToCreate{
				ContactPerson: buildPerson(e.GivenName, e.FamilyName, e.Email, e.Phone, e.OrgName, e.OrgTitle),
			})
		}

		result, err := srv.People.BatchCreateContacts(batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Batch Contacts Created")
		rb.KeyValue("Created", len(result.CreatedPeople))
		rb.Blank()
		for _, cp := range result.CreatedPeople {
			if cp.Person != nil {
				cs := personToSummary(cp.Person)
				rb.Item("%s", formatContactLine(cs))
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- batch_update_contacts (complete) ---

type BatchUpdateContactsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Contacts  string `json:"contacts" jsonschema:"required" jsonschema_description:"JSON object mapping resource names to update data. Each value has given_name family_name email phone org_name org_title etag."`
}

type ContactUpdate struct {
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	OrgName    string `json:"org_name"`
	OrgTitle   string `json:"org_title"`
	ETag       string `json:"etag"`
}

func createBatchUpdateContactsHandler(factory *services.Factory) mcp.ToolHandlerFor[BatchUpdateContactsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchUpdateContactsInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		var updates map[string]ContactUpdate
		if err := json.Unmarshal([]byte(input.Contacts), &updates); err != nil {
			return nil, nil, fmt.Errorf("invalid contacts JSON - provide object mapping resource_name to {given_name, family_name, email, etag}: %w", err)
		}

		batchReq := &people.BatchUpdateContactsRequest{
			Contacts:   make(map[string]people.Person),
			UpdateMask: "names,emailAddresses,phoneNumbers,organizations",
		}
		for rn, u := range updates {
			p := *buildPerson(u.GivenName, u.FamilyName, u.Email, u.Phone, u.OrgName, u.OrgTitle)
			p.Etag = u.ETag
			batchReq.Contacts[rn] = p
		}

		result, err := srv.People.BatchUpdateContacts(batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Batch Contacts Updated")
		rb.KeyValue("Updated", len(result.UpdateResult))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- batch_delete_contacts (complete) ---

type BatchDeleteContactsInput struct {
	UserEmail     string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ResourceNames []string `json:"resource_names" jsonschema:"required" jsonschema_description:"Resource names of contacts to delete (e.g. people/c12345)"`
}

func createBatchDeleteContactsHandler(factory *services.Factory) mcp.ToolHandlerFor[BatchDeleteContactsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchDeleteContactsInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		batchReq := &people.BatchDeleteContactsRequest{
			ResourceNames: input.ResourceNames,
		}

		_, err = srv.People.BatchDeleteContacts(batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Batch Contacts Deleted")
		rb.KeyValue("Deleted", len(input.ResourceNames))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- create_contact_group (complete) ---

type CreateContactGroupInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Name      string `json:"name" jsonschema:"required" jsonschema_description:"Name for the new contact group"`
}

func createCreateContactGroupHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateContactGroupInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateContactGroupInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		group := &people.CreateContactGroupRequest{
			ContactGroup: &people.ContactGroup{
				Name: input.Name,
			},
		}

		created, err := srv.ContactGroups.Create(group).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Contact Group Created")
		rb.KeyValue("Name", created.Name)
		rb.KeyValue("Resource Name", created.ResourceName)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- update_contact_group (complete) ---

type UpdateContactGroupInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ResourceName string `json:"resource_name" jsonschema:"required" jsonschema_description:"Resource name of the contact group (e.g. contactGroups/abc123)"`
	Name         string `json:"name" jsonschema:"required" jsonschema_description:"New name for the contact group"`
}

func createUpdateContactGroupHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdateContactGroupInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateContactGroupInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		group := &people.UpdateContactGroupRequest{
			ContactGroup: &people.ContactGroup{
				Name: input.Name,
			},
		}

		updated, err := srv.ContactGroups.Update(input.ResourceName, group).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Contact Group Updated")
		rb.KeyValue("Name", updated.Name)
		rb.KeyValue("Resource Name", updated.ResourceName)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- delete_contact_group (complete) ---

type DeleteContactGroupInput struct {
	UserEmail      string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ResourceName   string `json:"resource_name" jsonschema:"required" jsonschema_description:"Resource name of the contact group to delete"`
	DeleteContacts bool   `json:"delete_contacts,omitempty" jsonschema_description:"Also delete contacts in this group (default false)"`
}

func createDeleteContactGroupHandler(factory *services.Factory) mcp.ToolHandlerFor[DeleteContactGroupInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteContactGroupInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		call := srv.ContactGroups.Delete(input.ResourceName).Context(ctx)
		if input.DeleteContacts {
			call = call.DeleteContacts(true)
		}

		_, err = call.Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Contact Group Deleted")
		rb.KeyValue("Resource Name", input.ResourceName)
		if input.DeleteContacts {
			rb.Line("Associated contacts were also deleted.")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- modify_contact_group_members (complete) ---

type ModifyGroupMembersInput struct {
	UserEmail           string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ResourceName        string   `json:"resource_name" jsonschema:"required" jsonschema_description:"Resource name of the contact group"`
	AddResourceNames    []string `json:"add_resource_names,omitempty" jsonschema_description:"Resource names of contacts to add to the group"`
	RemoveResourceNames []string `json:"remove_resource_names,omitempty" jsonschema_description:"Resource names of contacts to remove from the group"`
}

func createModifyGroupMembersHandler(factory *services.Factory) mcp.ToolHandlerFor[ModifyGroupMembersInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ModifyGroupMembersInput) (*mcp.CallToolResult, any, error) {
		if len(input.AddResourceNames) == 0 && len(input.RemoveResourceNames) == 0 {
			return nil, nil, fmt.Errorf("specify at least one of add_resource_names or remove_resource_names")
		}

		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		modReq := &people.ModifyContactGroupMembersRequest{
			ResourceNamesToAdd:    input.AddResourceNames,
			ResourceNamesToRemove: input.RemoveResourceNames,
		}

		result, err := srv.ContactGroups.Members.Modify(input.ResourceName, modReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Contact Group Members Modified")
		rb.KeyValue("Group", input.ResourceName)
		if len(input.AddResourceNames) > 0 {
			rb.KeyValue("Added", len(input.AddResourceNames))
		}
		if len(input.RemoveResourceNames) > 0 {
			rb.KeyValue("Removed", len(input.RemoveResourceNames))
		}
		_ = result
		if len(result.NotFoundResourceNames) > 0 {
			rb.KeyValue("Not Found", strings.Join(result.NotFoundResourceNames, ", "))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

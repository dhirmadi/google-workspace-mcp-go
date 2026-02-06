package contacts

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- search_contacts (core) ---

type SearchContactsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Query     string `json:"query" jsonschema:"required" jsonschema_description:"Search query (name or email)"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum results (default 10)"`
}

type SearchContactsOutput struct {
	Contacts []ContactSummary `json:"contacts"`
}

func createSearchContactsHandler(factory *services.Factory) mcp.ToolHandlerFor[SearchContactsInput, SearchContactsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchContactsInput) (*mcp.CallToolResult, SearchContactsOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 10
		}

		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, SearchContactsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		result, err := srv.People.SearchContacts().
			Query(input.Query).
			ReadMask(personFieldsForList()).
			PageSize(int64(input.PageSize)).
			Context(ctx).
			Do()
		if err != nil {
			return nil, SearchContactsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		contacts := make([]ContactSummary, 0, len(result.Results))
		rb := response.New()
		rb.Header("Contact Search Results")
		rb.KeyValue("Query", input.Query)
		rb.KeyValue("Results", len(result.Results))
		rb.Blank()

		for _, r := range result.Results {
			cs := personToSummary(r.Person)
			contacts = append(contacts, cs)
			rb.Item("%s", formatContactLine(cs))
			rb.Line("    Resource: %s", cs.ResourceName)
			if cs.Organization != "" {
				rb.Line("    Org: %s", cs.Organization)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, SearchContactsOutput{Contacts: contacts}, nil
	}
}

// --- get_contact (core) ---

type GetContactInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ResourceName string `json:"resource_name" jsonschema:"required" jsonschema_description:"The contact resource name (e.g. people/c1234567890)"`
}

type GetContactOutput struct {
	Contact ContactSummary `json:"contact"`
}

func createGetContactHandler(factory *services.Factory) mcp.ToolHandlerFor[GetContactInput, GetContactOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetContactInput) (*mcp.CallToolResult, GetContactOutput, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, GetContactOutput{}, middleware.HandleGoogleAPIError(err)
		}

		person, err := srv.People.Get(input.ResourceName).
			PersonFields(personFieldsForRead()).
			Context(ctx).
			Do()
		if err != nil {
			return nil, GetContactOutput{}, middleware.HandleGoogleAPIError(err)
		}

		cs := personToSummary(person)

		rb := response.New()
		rb.Header("Contact Details")
		rb.KeyValue("Name", cs.DisplayName)
		rb.KeyValue("Resource", cs.ResourceName)
		for _, e := range cs.Emails {
			rb.KeyValue("Email", e)
		}
		for _, p := range cs.Phones {
			rb.KeyValue("Phone", p)
		}
		if cs.Organization != "" {
			rb.KeyValue("Organization", cs.Organization)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, GetContactOutput{Contact: cs}, nil
	}
}

// --- list_contacts (core) ---

type ListContactsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum contacts to return (default 25)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for pagination"`
}

type ListContactsOutput struct {
	Contacts      []ContactSummary `json:"contacts"`
	NextPageToken string           `json:"next_page_token,omitempty"`
}

func createListContactsHandler(factory *services.Factory) mcp.ToolHandlerFor[ListContactsInput, ListContactsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListContactsInput) (*mcp.CallToolResult, ListContactsOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 25
		}

		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, ListContactsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		call := srv.People.Connections.List("people/me").
			PersonFields(personFieldsForList()).
			PageSize(int64(input.PageSize)).
			SortOrder("FIRST_NAME_ASCENDING").
			Context(ctx)

		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, ListContactsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		contacts := make([]ContactSummary, 0, len(result.Connections))
		rb := response.New()
		rb.Header("Contacts")
		rb.KeyValue("Count", len(result.Connections))
		rb.KeyValue("Total", result.TotalPeople)
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()

		for _, p := range result.Connections {
			cs := personToSummary(p)
			contacts = append(contacts, cs)
			rb.Item("%s", formatContactLine(cs))
			if cs.Organization != "" {
				rb.Line("    Org: %s", cs.Organization)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ListContactsOutput{Contacts: contacts, NextPageToken: result.NextPageToken}, nil
	}
}

// --- create_contact (core) ---

type CreateContactInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	GivenName  string `json:"given_name" jsonschema:"required" jsonschema_description:"First name"`
	FamilyName string `json:"family_name,omitempty" jsonschema_description:"Last name"`
	Email      string `json:"email,omitempty" jsonschema_description:"Email address"`
	Phone      string `json:"phone,omitempty" jsonschema_description:"Phone number"`
	OrgName    string `json:"organization,omitempty" jsonschema_description:"Organization name"`
	OrgTitle   string `json:"job_title,omitempty" jsonschema_description:"Job title"`
}

func createCreateContactHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateContactInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateContactInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		person := buildPerson(input.GivenName, input.FamilyName, input.Email, input.Phone, input.OrgName, input.OrgTitle)

		created, err := srv.People.CreateContact(person).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		cs := personToSummary(created)
		rb := response.New()
		rb.Header("Contact Created")
		rb.KeyValue("Name", cs.DisplayName)
		rb.KeyValue("Resource", cs.ResourceName)
		if len(cs.Emails) > 0 {
			rb.KeyValue("Email", cs.Emails[0])
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- update_contact (extended) ---

type UpdateContactInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ResourceName string `json:"resource_name" jsonschema:"required" jsonschema_description:"The contact resource name"`
	GivenName    string `json:"given_name,omitempty" jsonschema_description:"Updated first name"`
	FamilyName   string `json:"family_name,omitempty" jsonschema_description:"Updated last name"`
	Email        string `json:"email,omitempty" jsonschema_description:"Updated email address"`
	Phone        string `json:"phone,omitempty" jsonschema_description:"Updated phone number"`
	OrgName      string `json:"organization,omitempty" jsonschema_description:"Updated organization"`
	OrgTitle     string `json:"job_title,omitempty" jsonschema_description:"Updated job title"`
}

func createUpdateContactHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdateContactInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateContactInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		// Get existing contact for etag
		existing, err := srv.People.Get(input.ResourceName).
			PersonFields(personFieldsForRead()).
			Context(ctx).
			Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		person := buildPerson(input.GivenName, input.FamilyName, input.Email, input.Phone, input.OrgName, input.OrgTitle)
		person.Etag = existing.Etag

		updateFields := "names,emailAddresses,phoneNumbers,organizations"

		updated, err := srv.People.UpdateContact(input.ResourceName, person).
			UpdatePersonFields(updateFields).
			Context(ctx).
			Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		cs := personToSummary(updated)
		rb := response.New()
		rb.Header("Contact Updated")
		rb.KeyValue("Name", cs.DisplayName)
		rb.KeyValue("Resource", cs.ResourceName)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- delete_contact (extended) ---

type DeleteContactInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ResourceName string `json:"resource_name" jsonschema:"required" jsonschema_description:"The contact resource name to delete"`
}

func createDeleteContactHandler(factory *services.Factory) mcp.ToolHandlerFor[DeleteContactInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteContactInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		_, err = srv.People.DeleteContact(input.ResourceName).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Contact Deleted")
		rb.KeyValue("Resource", input.ResourceName)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- list_contact_groups (extended) ---

type ListContactGroupsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum results (default 25)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for pagination"`
}

type ListContactGroupsOutput struct {
	Groups        []ContactGroupSummary `json:"groups"`
	NextPageToken string                `json:"next_page_token,omitempty"`
}

func createListContactGroupsHandler(factory *services.Factory) mcp.ToolHandlerFor[ListContactGroupsInput, ListContactGroupsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListContactGroupsInput) (*mcp.CallToolResult, ListContactGroupsOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 25
		}

		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, ListContactGroupsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		call := srv.ContactGroups.List().
			PageSize(int64(input.PageSize)).
			Context(ctx)

		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, ListContactGroupsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		groups := make([]ContactGroupSummary, 0, len(result.ContactGroups))
		rb := response.New()
		rb.Header("Contact Groups")
		rb.KeyValue("Count", len(result.ContactGroups))
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()

		for _, g := range result.ContactGroups {
			gs := contactGroupToSummary(g)
			groups = append(groups, gs)
			rb.Item("%s (%s)", gs.Name, gs.GroupType)
			rb.Line("    Resource: %s | Members: %d", gs.ResourceName, gs.MemberCount)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ListContactGroupsOutput{Groups: groups, NextPageToken: result.NextPageToken}, nil
	}
}

// --- get_contact_group (extended) ---

type GetContactGroupInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ResourceName string `json:"resource_name" jsonschema:"required" jsonschema_description:"The contact group resource name (e.g. contactGroups/123)"`
}

type GetContactGroupOutput struct {
	Group ContactGroupSummary `json:"group"`
}

func createGetContactGroupHandler(factory *services.Factory) mcp.ToolHandlerFor[GetContactGroupInput, GetContactGroupOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetContactGroupInput) (*mcp.CallToolResult, GetContactGroupOutput, error) {
		srv, err := factory.People(ctx, input.UserEmail)
		if err != nil {
			return nil, GetContactGroupOutput{}, middleware.HandleGoogleAPIError(err)
		}

		group, err := srv.ContactGroups.Get(input.ResourceName).Context(ctx).Do()
		if err != nil {
			return nil, GetContactGroupOutput{}, middleware.HandleGoogleAPIError(err)
		}

		gs := contactGroupToSummary(group)

		rb := response.New()
		rb.Header("Contact Group")
		rb.KeyValue("Name", gs.Name)
		rb.KeyValue("Resource", gs.ResourceName)
		rb.KeyValue("Members", gs.MemberCount)
		rb.KeyValue("Type", gs.GroupType)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, GetContactGroupOutput{Group: gs}, nil
	}
}


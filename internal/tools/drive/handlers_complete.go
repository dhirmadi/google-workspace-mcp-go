package drive

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- get_drive_file_permissions (complete) ---

type GetFilePermissionsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID    string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID"`
}

type GetFilePermissionsOutput struct {
	FileID      string             `json:"file_id"`
	Permissions []PermissionDetail `json:"permissions"`
}

type PermissionDetail struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Type        string `json:"type"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Domain      string `json:"domain,omitempty"`
}

func createGetFilePermissionsHandler(factory *services.Factory) mcp.ToolHandlerFor[GetFilePermissionsInput, GetFilePermissionsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetFilePermissionsInput) (*mcp.CallToolResult, GetFilePermissionsOutput, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, GetFilePermissionsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		result, err := srv.Permissions.List(input.FileID).
			Fields("permissions(id, role, type, emailAddress, displayName, domain)").
			Context(ctx).
			Do()
		if err != nil {
			return nil, GetFilePermissionsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		perms := make([]PermissionDetail, 0, len(result.Permissions))
		rb := response.New()
		rb.Header("File Permissions")
		rb.KeyValue("File ID", input.FileID)
		rb.KeyValue("Permissions", len(result.Permissions))
		rb.Blank()

		for _, p := range result.Permissions {
			pd := PermissionDetail{
				ID:          p.Id,
				Role:        p.Role,
				Type:        p.Type,
				Email:       p.EmailAddress,
				DisplayName: p.DisplayName,
				Domain:      p.Domain,
			}
			perms = append(perms, pd)

			rb.Item("[%s] %s", p.Role, p.Type)
			if p.EmailAddress != "" {
				rb.Line("    Email: %s", p.EmailAddress)
			}
			if p.DisplayName != "" {
				rb.Line("    Name: %s", p.DisplayName)
			}
			if p.Domain != "" {
				rb.Line("    Domain: %s", p.Domain)
			}
			rb.Line("    ID: %s", p.Id)
		}

		return rb.TextResult(), GetFilePermissionsOutput{FileID: input.FileID, Permissions: perms}, nil
	}
}

// --- check_drive_file_public_access (complete) ---

type CheckPublicAccessInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID    string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID"`
}

type CheckPublicAccessOutput struct {
	FileID      string `json:"file_id"`
	IsPublic    bool   `json:"is_public"`
	PublicRole  string `json:"public_role,omitempty"`
	DomainShare bool   `json:"domain_share"`
	DomainRole  string `json:"domain_role,omitempty"`
}

func createCheckPublicAccessHandler(factory *services.Factory) mcp.ToolHandlerFor[CheckPublicAccessInput, CheckPublicAccessOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CheckPublicAccessInput) (*mcp.CallToolResult, CheckPublicAccessOutput, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, CheckPublicAccessOutput{}, middleware.HandleGoogleAPIError(err)
		}

		result, err := srv.Permissions.List(input.FileID).
			Fields("permissions(id, role, type, domain)").
			Context(ctx).
			Do()
		if err != nil {
			return nil, CheckPublicAccessOutput{}, middleware.HandleGoogleAPIError(err)
		}

		output := CheckPublicAccessOutput{
			FileID: input.FileID,
		}

		rb := response.New()
		rb.Header("Public Access Check")
		rb.KeyValue("File ID", input.FileID)

		for _, p := range result.Permissions {
			if p.Type == "anyone" {
				output.IsPublic = true
				output.PublicRole = p.Role
			}
			if p.Type == "domain" {
				output.DomainShare = true
				output.DomainRole = p.Role
			}
		}

		if output.IsPublic {
			rb.KeyValue("Public", "YES — accessible to anyone with the link")
			rb.KeyValue("Public Role", output.PublicRole)
		} else {
			rb.KeyValue("Public", "NO — not publicly accessible")
		}

		if output.DomainShare {
			rb.KeyValue("Domain Shared", "YES")
			rb.KeyValue("Domain Role", output.DomainRole)
		}

		return rb.TextResult(), output, nil
	}
}

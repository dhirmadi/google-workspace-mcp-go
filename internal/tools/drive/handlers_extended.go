package drive

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/drive/v3"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/validate"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- list_drive_items (extended) ---

type ListDriveItemsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FolderID  string `json:"folder_id,omitempty" jsonschema_description:"Folder ID to list (default: root)"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum results (default 25)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for pagination"`
}

type ListDriveItemsOutput struct {
	Files         []FileSummary `json:"files"`
	NextPageToken string        `json:"next_page_token,omitempty"`
}

func createListDriveItemsHandler(factory *services.Factory) mcp.ToolHandlerFor[ListDriveItemsInput, ListDriveItemsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListDriveItemsInput) (*mcp.CallToolResult, ListDriveItemsOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 25
		}

		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, ListDriveItemsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		folderID := input.FolderID
		if folderID == "" {
			folderID = "root"
		}
		if err := validate.DriveID(folderID); err != nil {
			return nil, ListDriveItemsOutput{}, err
		}

		q := fmt.Sprintf("'%s' in parents and trashed=false", folderID)

		call := srv.Files.List().
			Q(q).
			PageSize(int64(input.PageSize)).
			Fields("nextPageToken, files(id, name, mimeType, size, modifiedTime, webViewLink)").
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			OrderBy("folder,name").
			Context(ctx)

		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, ListDriveItemsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		files := make([]FileSummary, 0, len(result.Files))
		rb := response.New()
		rb.Header("Drive Items")
		rb.KeyValue("Folder", folderID)
		rb.KeyValue("Count", len(result.Files))
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()

		for _, f := range result.Files {
			fs := fileToSummary(f)
			files = append(files, fs)
			rb.Item("%s (%s)", fs.Name, formatFileType(fs.MimeType))
			rb.Line("    ID: %s", fs.ID)
		}

		return rb.TextResult(), ListDriveItemsOutput{Files: files, NextPageToken: result.NextPageToken}, nil
	}
}

// --- copy_drive_file (extended) ---

type CopyFileInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID    string `json:"file_id" jsonschema:"required" jsonschema_description:"The file ID to copy"`
	Name      string `json:"name,omitempty" jsonschema_description:"Name for the copy (default: Copy of original)"`
	FolderID  string `json:"folder_id,omitempty" jsonschema_description:"Destination folder ID"`
}

func createCopyFileHandler(factory *services.Factory) mcp.ToolHandlerFor[CopyFileInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CopyFileInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		copyFile := &drive.File{}
		if input.Name != "" {
			copyFile.Name = input.Name
		}
		if input.FolderID != "" {
			copyFile.Parents = []string{input.FolderID}
		}

		created, err := srv.Files.Copy(input.FileID, copyFile).
			Fields("id, name, mimeType, webViewLink").
			SupportsAllDrives(true).
			Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("File Copied")
		rb.KeyValue("Name", created.Name)
		rb.KeyValue("ID", created.Id)
		rb.KeyValue("Type", formatFileType(created.MimeType))
		if created.WebViewLink != "" {
			rb.KeyValue("Link", created.WebViewLink)
		}

		return rb.TextResult(), nil, nil
	}
}

// --- update_drive_file (extended) ---

type UpdateFileInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID       string `json:"file_id" jsonschema:"required" jsonschema_description:"The file ID to update"`
	Name         string `json:"name,omitempty" jsonschema_description:"New file name"`
	Content      string `json:"content,omitempty" jsonschema_description:"New text content (replaces file content)"`
	MoveToFolder string `json:"move_to_folder,omitempty" jsonschema_description:"Folder ID to move file to"`
}

func createUpdateFileHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdateFileInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateFileInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		fileMetadata := &drive.File{}
		if input.Name != "" {
			fileMetadata.Name = input.Name
		}

		call := srv.Files.Update(input.FileID, fileMetadata).
			SupportsAllDrives(true).
			Fields("id, name, mimeType, webViewLink").
			Context(ctx)

		if input.Content != "" {
			call = call.Media(strings.NewReader(input.Content))
		}

		if input.MoveToFolder != "" {
			// Get current parents to remove
			existing, err := srv.Files.Get(input.FileID).Fields("parents").SupportsAllDrives(true).Context(ctx).Do()
			if err != nil {
				return nil, nil, middleware.HandleGoogleAPIError(err)
			}
			call = call.AddParents(input.MoveToFolder)
			if len(existing.Parents) > 0 {
				call = call.RemoveParents(strings.Join(existing.Parents, ","))
			}
		}

		updated, err := call.Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("File Updated")
		rb.KeyValue("Name", updated.Name)
		rb.KeyValue("ID", updated.Id)
		if updated.WebViewLink != "" {
			rb.KeyValue("Link", updated.WebViewLink)
		}

		return rb.TextResult(), nil, nil
	}
}

// --- update_drive_permission (extended) ---

type UpdatePermissionInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID       string `json:"file_id" jsonschema:"required" jsonschema_description:"The file ID"`
	PermissionID string `json:"permission_id" jsonschema:"required" jsonschema_description:"The permission ID to update"`
	Role         string `json:"role" jsonschema:"required" jsonschema_description:"New role: reader/writer/commenter"`
}

func createUpdatePermissionHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdatePermissionInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdatePermissionInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		updated, err := srv.Permissions.Update(input.FileID, input.PermissionID, &drive.Permission{
			Role: input.Role,
		}).SupportsAllDrives(true).
			Fields("id, type, role, emailAddress").
			Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Permission Updated")
		rb.KeyValue("File ID", input.FileID)
		rb.KeyValue("Permission", formatPermission(updated))

		return rb.TextResult(), nil, nil
	}
}

// --- remove_drive_permission (extended) ---

type RemovePermissionInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID       string `json:"file_id" jsonschema:"required" jsonschema_description:"The file ID"`
	PermissionID string `json:"permission_id" jsonschema:"required" jsonschema_description:"The permission ID to remove"`
}

func createRemovePermissionHandler(factory *services.Factory) mcp.ToolHandlerFor[RemovePermissionInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RemovePermissionInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		err = srv.Permissions.Delete(input.FileID, input.PermissionID).
			SupportsAllDrives(true).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Permission Removed")
		rb.KeyValue("File ID", input.FileID)
		rb.KeyValue("Permission ID", input.PermissionID)

		return rb.TextResult(), nil, nil
	}
}

// --- transfer_drive_ownership (extended) ---

type TransferOwnershipInput struct {
	UserEmail     string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID        string `json:"file_id" jsonschema:"required" jsonschema_description:"The file ID to transfer"`
	NewOwnerEmail string `json:"new_owner_email" jsonschema:"required" jsonschema_description:"Email of the new owner"`
}

func createTransferOwnershipHandler(factory *services.Factory) mcp.ToolHandlerFor[TransferOwnershipInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input TransferOwnershipInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		_, err = srv.Permissions.Create(input.FileID, &drive.Permission{
			Type:         "user",
			Role:         "owner",
			EmailAddress: input.NewOwnerEmail,
		}).TransferOwnership(true).
			SupportsAllDrives(true).
			Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Ownership Transferred")
		rb.KeyValue("File ID", input.FileID)
		rb.KeyValue("New Owner", input.NewOwnerEmail)

		return rb.TextResult(), nil, nil
	}
}

// --- batch_share_drive_file (extended) ---

type BatchShareInput struct {
	UserEmail        string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileIDs          []string `json:"file_ids" jsonschema:"required" jsonschema_description:"File IDs to share"`
	ShareWith        string   `json:"share_with" jsonschema:"required" jsonschema_description:"Email address to share with"`
	Role             string   `json:"role,omitempty" jsonschema_description:"Permission role: reader/writer/commenter (default reader)"`
	SendNotification bool     `json:"send_notification,omitempty" jsonschema_description:"Send notification email (default true)"`
}

func createBatchShareHandler(factory *services.Factory) mcp.ToolHandlerFor[BatchShareInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchShareInput) (*mcp.CallToolResult, any, error) {
		if len(input.FileIDs) == 0 {
			return nil, nil, fmt.Errorf("file_ids cannot be empty")
		}

		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		if input.Role == "" {
			input.Role = "reader"
		}

		total := len(input.FileIDs)
		shared := 0
		var errors []string

		for i, fileID := range input.FileIDs {
			// Report progress
			if pt := req.Params.GetProgressToken(); pt != nil {
				_ = req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
					ProgressToken: pt,
					Progress:      float64(i),
					Total:         float64(total),
					Message:       fmt.Sprintf("Sharing file %d/%d", i+1, total),
				})
			}

			_, err := srv.Permissions.Create(fileID, &drive.Permission{
				Type:         "user",
				Role:         input.Role,
				EmailAddress: input.ShareWith,
			}).SupportsAllDrives(true).
				SendNotificationEmail(input.SendNotification).
				Context(ctx).Do()
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", fileID, err))
				continue
			}
			shared++
		}

		rb := response.New()
		rb.Header("Batch Share Complete")
		rb.KeyValue("Shared with", input.ShareWith)
		rb.KeyValue("Role", input.Role)
		rb.KeyValue("Successful", shared)
		rb.KeyValue("Failed", len(errors))
		if len(errors) > 0 {
			rb.Blank()
			rb.Section("Errors")
			for _, e := range errors {
				rb.Item("%s", e)
			}
		}

		return rb.TextResult(), nil, nil
	}
}

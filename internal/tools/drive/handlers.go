package drive

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/drive/v3"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/office"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- search_drive_files ---

type SearchFilesInput struct {
	UserEmail           string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Query               string `json:"query" jsonschema:"required" jsonschema_description:"Google Drive search query using Drive query syntax"`
	PageSize            int    `json:"page_size,omitempty" jsonschema_description:"Maximum number of results to return (default 10)"`
	DriveID             string `json:"drive_id,omitempty" jsonschema_description:"ID of a shared drive to search within"`
	IncludeSharedDrives bool   `json:"include_items_from_all_drives,omitempty" jsonschema_description:"Include shared drive items in results (default true)"`
}

type SearchFilesOutput struct {
	Files         []FileSummary `json:"files"`
	Query         string        `json:"query"`
	NextPageToken string        `json:"next_page_token,omitempty"`
	ResultCount   int           `json:"result_count"`
}

func createSearchFilesHandler(factory *services.Factory) mcp.ToolHandlerFor[SearchFilesInput, SearchFilesOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchFilesInput) (*mcp.CallToolResult, SearchFilesOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 10
		}

		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, SearchFilesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		call := srv.Files.List().
			Q(input.Query).
			PageSize(int64(input.PageSize)).
			Fields("nextPageToken, files(id, name, mimeType, size, modifiedTime, webViewLink)").
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			Context(ctx)

		if input.DriveID != "" {
			call = call.DriveId(input.DriveID).Corpora("drive")
		}

		result, err := call.Do()
		if err != nil {
			return nil, SearchFilesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		files := make([]FileSummary, 0, len(result.Files))
		for _, f := range result.Files {
			files = append(files, fileToSummary(f))
		}

		rb := response.New()
		rb.Header("Drive Search Results")
		rb.KeyValue("Query", input.Query)
		rb.KeyValue("Results", len(files))
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()
		for _, f := range files {
			rb.Item("%s (%s)", f.Name, formatFileType(f.MimeType))
			size := formatSize(f.Size)
			if size != "" {
				rb.Line("    Size: %s | Modified: %s", size, f.ModifiedTime)
			} else {
				rb.Line("    Modified: %s", f.ModifiedTime)
			}
			rb.Line("    ID: %s", f.ID)
			if f.WebViewLink != "" {
				rb.Line("    Link: %s", f.WebViewLink)
			}
		}

		output := SearchFilesOutput{
			Files:         files,
			Query:         input.Query,
			NextPageToken: result.NextPageToken,
			ResultCount:   len(files),
		}

		return rb.TextResult(), output, nil
	}
}

// --- get_drive_file_content ---

type GetFileContentInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID    string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID"`
}

type GetFileContentOutput struct {
	Content  string `json:"content"`
	Title    string `json:"title"`
	MimeType string `json:"mime_type"`
}

func createGetFileContentHandler(factory *services.Factory) mcp.ToolHandlerFor[GetFileContentInput, GetFileContentOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetFileContentInput) (*mcp.CallToolResult, GetFileContentOutput, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, GetFileContentOutput{}, middleware.HandleGoogleAPIError(err)
		}

		// Get file metadata first
		file, err := srv.Files.Get(input.FileID).
			Fields("id, name, mimeType, size").
			SupportsAllDrives(true).
			Context(ctx).
			Do()
		if err != nil {
			return nil, GetFileContentOutput{}, middleware.HandleGoogleAPIError(err)
		}

		var content string

		if isGoogleNativeType(file.MimeType) {
			// Export Google native files
			exportMime := mimeTypeForExport(file.MimeType)
			if exportMime == "" {
				return nil, GetFileContentOutput{}, fmt.Errorf("unsupported Google file type %q for text export", file.MimeType)
			}
			resp, err := srv.Files.Export(input.FileID, exportMime).Context(ctx).Download()
			if err != nil {
				return nil, GetFileContentOutput{}, middleware.HandleGoogleAPIError(err)
			}
			defer resp.Body.Close()
			data, err := io.ReadAll(io.LimitReader(resp.Body, office.MaxFileSize))
			if err != nil {
				return nil, GetFileContentOutput{}, fmt.Errorf("reading exported content: %w", err)
			}
			content = string(data)
		} else {
			// Download binary files
			resp, err := srv.Files.Get(input.FileID).
				SupportsAllDrives(true).
				Context(ctx).
				Download()
			if err != nil {
				return nil, GetFileContentOutput{}, middleware.HandleGoogleAPIError(err)
			}
			defer resp.Body.Close()
			data, err := io.ReadAll(io.LimitReader(resp.Body, office.MaxFileSize))
			if err != nil {
				return nil, GetFileContentOutput{}, fmt.Errorf("reading file content: %w", err)
			}

			// Try Office XML extraction
			if isOfficeType(file.MimeType) {
				extracted, extractErr := office.ExtractText(data, file.MimeType)
				if extractErr == nil {
					content = extracted
				} else {
					content = string(data)
				}
			} else {
				// Try UTF-8 decode
				content = string(data)
			}
		}

		rb := response.New()
		rb.Header("Drive File Content")
		rb.KeyValue("Title", file.Name)
		rb.KeyValue("Type", formatFileType(file.MimeType))
		rb.KeyValue("ID", file.Id)
		rb.Blank()
		rb.Raw(content)

		return rb.TextResult(), GetFileContentOutput{Content: content, Title: file.Name, MimeType: file.MimeType}, nil
	}
}

// --- get_drive_file_download_url ---

type GetDownloadURLInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID       string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID"`
	ExportFormat string `json:"export_format,omitempty" jsonschema_description:"Export format for Google native files (pdf/docx/xlsx/csv/pptx)"`
}

type GetDownloadURLOutput struct {
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
	MimeType    string `json:"mime_type"`
}

func createGetDownloadURLHandler(factory *services.Factory) mcp.ToolHandlerFor[GetDownloadURLInput, GetDownloadURLOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetDownloadURLInput) (*mcp.CallToolResult, GetDownloadURLOutput, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, GetDownloadURLOutput{}, middleware.HandleGoogleAPIError(err)
		}

		file, err := srv.Files.Get(input.FileID).
			Fields("id, name, mimeType, webContentLink").
			SupportsAllDrives(true).
			Context(ctx).
			Do()
		if err != nil {
			return nil, GetDownloadURLOutput{}, middleware.HandleGoogleAPIError(err)
		}

		var downloadURL string
		if isGoogleNativeType(file.MimeType) {
			exportMime := mimeTypeForDownloadURL(file.MimeType)
			if input.ExportFormat != "" {
				exportMime = exportFormatToMime(input.ExportFormat)
			}
			if exportMime == "" {
				return nil, GetDownloadURLOutput{}, fmt.Errorf("unsupported export format for %q", file.MimeType)
			}
			downloadURL = fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s/export?mimeType=%s", input.FileID, exportMime)
		} else {
			downloadURL = file.WebContentLink
			if downloadURL == "" {
				downloadURL = fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", input.FileID)
			}
		}

		rb := response.New()
		rb.Header("Drive File Download URL")
		rb.KeyValue("File", file.Name)
		rb.KeyValue("Type", formatFileType(file.MimeType))
		rb.KeyValue("Download URL", downloadURL)

		return rb.TextResult(), GetDownloadURLOutput{DownloadURL: downloadURL, FileName: file.Name, MimeType: file.MimeType}, nil
	}
}

// --- create_drive_file ---

type CreateFileInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileName  string `json:"file_name" jsonschema:"required" jsonschema_description:"Name for the new file"`
	Content   string `json:"content,omitempty" jsonschema_description:"Text content to write to the file"`
	FolderID  string `json:"folder_id,omitempty" jsonschema_description:"ID of the parent folder (default: root)"`
	MimeType  string `json:"mime_type,omitempty" jsonschema_description:"MIME type of the file (default: text/plain)"`
}

func createCreateFileHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateFileInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateFileInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		if input.MimeType == "" {
			input.MimeType = "text/plain"
		}

		fileMetadata := &drive.File{
			Name:     input.FileName,
			MimeType: input.MimeType,
		}
		if input.FolderID != "" {
			fileMetadata.Parents = []string{input.FolderID}
		}

		var created *drive.File
		if input.Content != "" {
			created, err = srv.Files.Create(fileMetadata).
				Media(strings.NewReader(input.Content)).
				Fields("id, name, mimeType, webViewLink").
				SupportsAllDrives(true).
				Context(ctx).
				Do()
		} else {
			created, err = srv.Files.Create(fileMetadata).
				Fields("id, name, mimeType, webViewLink").
				SupportsAllDrives(true).
				Context(ctx).
				Do()
		}
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("File Created")
		rb.KeyValue("Name", created.Name)
		rb.KeyValue("ID", created.Id)
		rb.KeyValue("Type", formatFileType(created.MimeType))
		if created.WebViewLink != "" {
			rb.KeyValue("Link", created.WebViewLink)
		}

		return rb.TextResult(), nil, nil
	}
}

// --- import_to_google_doc ---

type ImportToDocInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID    string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID to import"`
	Title     string `json:"title,omitempty" jsonschema_description:"Title for the new Google Doc"`
	FolderID  string `json:"folder_id,omitempty" jsonschema_description:"Destination folder ID for the new Google Doc. If omitted, the file is created in the user's My Drive root."`
}

func createImportToDocHandler(factory *services.Factory) mcp.ToolHandlerFor[ImportToDocInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ImportToDocInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		// Get original file info
		original, err := srv.Files.Get(input.FileID).
			Fields("id, name, parents").
			SupportsAllDrives(true).
			Context(ctx).
			Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		title := input.Title
		if title == "" {
			title = original.Name
		}

		// Copy the file as a Google Doc
		copiedFile := &drive.File{
			Name:     title,
			MimeType: "application/vnd.google-apps.document",
		}
		if input.FolderID != "" {
			copiedFile.Parents = []string{input.FolderID}
		} else if len(original.Parents) > 0 {
			copiedFile.Parents = original.Parents
		}
		// Note: if neither folder_id nor original parents are available,
		// the file is created in the user's My Drive root by default.

		created, err := srv.Files.Copy(input.FileID, copiedFile).
			Fields("id, name, mimeType, webViewLink").
			SupportsAllDrives(true).
			Context(ctx).
			Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Imported to Google Doc")
		rb.KeyValue("Name", created.Name)
		rb.KeyValue("ID", created.Id)
		if created.WebViewLink != "" {
			rb.KeyValue("Link", created.WebViewLink)
		}

		return rb.TextResult(), nil, nil
	}
}

// --- share_drive_file ---

type ShareFileInput struct {
	UserEmail        string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID           string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID to share"`
	ShareWith        string `json:"share_with,omitempty" jsonschema_description:"Email address (for user/group) or domain name. Omit for 'anyone' sharing."`
	Role             string `json:"role,omitempty" jsonschema_description:"Permission role: reader/commenter/writer (default: reader)"`
	ShareType        string `json:"share_type,omitempty" jsonschema_description:"Type of sharing: user/group/domain/anyone (default: user)"`
	SendNotification bool   `json:"send_notification,omitempty" jsonschema_description:"Whether to send a notification email (default true)"`
	EmailMessage     string `json:"email_message,omitempty" jsonschema_description:"Custom message for the notification email"`
}

func createShareFileHandler(factory *services.Factory) mcp.ToolHandlerFor[ShareFileInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ShareFileInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		if input.Role == "" {
			input.Role = "reader"
		}
		if input.ShareType == "" {
			input.ShareType = "user"
		}

		perm := &drive.Permission{
			Type: input.ShareType,
			Role: input.Role,
		}
		if input.ShareWith != "" {
			if input.ShareType == "domain" {
				perm.Domain = input.ShareWith
			} else {
				perm.EmailAddress = input.ShareWith
			}
		}

		call := srv.Permissions.Create(input.FileID, perm).
			SupportsAllDrives(true).
			SendNotificationEmail(input.SendNotification).
			Fields("id, type, role, emailAddress, displayName, domain").
			Context(ctx)

		if input.EmailMessage != "" {
			call = call.EmailMessage(input.EmailMessage)
		}

		created, err := call.Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("File Shared")
		rb.KeyValue("File ID", input.FileID)
		rb.KeyValue("Permission", formatPermission(created))
		rb.KeyValue("Permission ID", created.Id)

		return rb.TextResult(), nil, nil
	}
}

// --- get_drive_shareable_link ---

type GetShareableLinkInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID    string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID"`
}

type GetShareableLinkOutput struct {
	WebViewLink string           `json:"web_view_link"`
	Permissions []PermissionInfo `json:"permissions"`
}

func createGetShareableLinkHandler(factory *services.Factory) mcp.ToolHandlerFor[GetShareableLinkInput, GetShareableLinkOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetShareableLinkInput) (*mcp.CallToolResult, GetShareableLinkOutput, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, GetShareableLinkOutput{}, middleware.HandleGoogleAPIError(err)
		}

		file, err := srv.Files.Get(input.FileID).
			Fields("id, name, webViewLink, permissions(id, type, role, emailAddress, displayName, domain)").
			SupportsAllDrives(true).
			Context(ctx).
			Do()
		if err != nil {
			return nil, GetShareableLinkOutput{}, middleware.HandleGoogleAPIError(err)
		}

		perms := make([]PermissionInfo, 0, len(file.Permissions))
		for _, p := range file.Permissions {
			perms = append(perms, permissionToInfo(p))
		}

		rb := response.New()
		rb.Header("Drive Shareable Link")
		rb.KeyValue("File", file.Name)
		rb.KeyValue("Link", file.WebViewLink)
		rb.Blank()
		rb.Section("Current Permissions")
		for _, p := range file.Permissions {
			rb.Item("%s", formatPermission(p))
		}

		return rb.TextResult(), GetShareableLinkOutput{WebViewLink: file.WebViewLink, Permissions: perms}, nil
	}
}

// exportFormatToMime converts a user-friendly export format to a MIME type.
func exportFormatToMime(format string) string {
	switch strings.ToLower(format) {
	case "pdf":
		return "application/pdf"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "csv":
		return "text/csv"
	case "pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return ""
	}
}

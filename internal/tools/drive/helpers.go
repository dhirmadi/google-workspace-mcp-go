package drive

import (
	"fmt"
	"strings"

	"google.golang.org/api/drive/v3"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/format"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/office"
)

// FileSummary is a compact representation of a Drive file.
type FileSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MimeType     string `json:"mime_type"`
	Size         int64  `json:"size,omitempty"`
	ModifiedTime string `json:"modified_time,omitempty"`
	WebViewLink  string `json:"web_view_link,omitempty"`
}

// PermissionInfo represents a sharing permission.
type PermissionInfo struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	EmailAddress string `json:"email_address,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	Domain       string `json:"domain,omitempty"`
}

// fileToSummary converts a Drive file to a compact summary.
func fileToSummary(f *drive.File) FileSummary {
	return FileSummary{
		ID:           f.Id,
		Name:         f.Name,
		MimeType:     f.MimeType,
		Size:         f.Size,
		ModifiedTime: f.ModifiedTime,
		WebViewLink:  f.WebViewLink,
	}
}

// formatFileType returns a human-readable file type from a MIME type.
func formatFileType(mimeType string) string {
	switch mimeType {
	case "application/vnd.google-apps.document":
		return "Google Doc"
	case "application/vnd.google-apps.spreadsheet":
		return "Google Sheet"
	case "application/vnd.google-apps.presentation":
		return "Google Slides"
	case "application/vnd.google-apps.folder":
		return "Folder"
	case "application/vnd.google-apps.form":
		return "Google Form"
	case "application/pdf":
		return "PDF"
	default:
		if strings.HasPrefix(mimeType, "image/") {
			return "Image"
		}
		if strings.HasPrefix(mimeType, "video/") {
			return "Video"
		}
		if strings.HasPrefix(mimeType, "audio/") {
			return "Audio"
		}
		return mimeType
	}
}

// formatSize returns a human-readable file size.
func formatSize(bytes int64) string {
	return format.ByteSize(bytes)
}

// permissionToInfo converts a Drive permission to a summary.
func permissionToInfo(p *drive.Permission) PermissionInfo {
	return PermissionInfo{
		ID:           p.Id,
		Type:         p.Type,
		Role:         p.Role,
		EmailAddress: p.EmailAddress,
		DisplayName:  p.DisplayName,
		Domain:       p.Domain,
	}
}

// formatPermission returns a human-readable description of a permission.
func formatPermission(p *drive.Permission) string {
	switch p.Type {
	case "user":
		return fmt.Sprintf("%s (%s) — %s", p.DisplayName, p.EmailAddress, p.Role)
	case "group":
		return fmt.Sprintf("Group: %s — %s", p.EmailAddress, p.Role)
	case "domain":
		return fmt.Sprintf("Domain: %s — %s", p.Domain, p.Role)
	case "anyone":
		return fmt.Sprintf("Anyone with the link — %s", p.Role)
	default:
		return fmt.Sprintf("%s: %s — %s", p.Type, p.EmailAddress, p.Role)
	}
}

// mimeTypeForExport returns the export MIME type for a Google Workspace file.
func mimeTypeForExport(googleMimeType string) string {
	switch googleMimeType {
	case "application/vnd.google-apps.document":
		return "text/plain"
	case "application/vnd.google-apps.spreadsheet":
		return "text/csv"
	case "application/vnd.google-apps.presentation":
		return "text/plain"
	case "application/vnd.google-apps.drawing":
		return "image/png"
	default:
		return ""
	}
}

// mimeTypeForDownloadURL returns the preferred download MIME type.
func mimeTypeForDownloadURL(googleMimeType string) string {
	switch googleMimeType {
	case "application/vnd.google-apps.document":
		return "application/pdf"
	case "application/vnd.google-apps.spreadsheet":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "application/vnd.google-apps.presentation":
		return "application/pdf"
	default:
		return ""
	}
}

// isGoogleNativeType returns true if the MIME type is a Google Workspace native type.
func isGoogleNativeType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "application/vnd.google-apps.")
}

// isOfficeType returns true if the MIME type is a Microsoft Office XML format.
func isOfficeType(mimeType string) bool {
	return office.IsOfficeType(mimeType)
}

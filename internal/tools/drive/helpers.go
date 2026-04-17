package drive

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
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

// driveBase64PayloadNonEmpty reports whether content_base64 carries non-whitespace payload.
func driveBase64PayloadNonEmpty(s string) bool {
	return normalizeBase64Payload(s) != ""
}

func normalizeBase64Payload(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}

// decodeDriveContentBase64 decodes standard or URL-safe base64 (RFC 4648) after stripping
// common whitespace from pasted payloads. Enforces maxDecodedBytes on the decoded length.
func decodeDriveContentBase64(s string, maxDecodedBytes int) ([]byte, error) {
	payload := normalizeBase64Payload(s)
	if payload == "" {
		return nil, fmt.Errorf("content_base64 is empty")
	}
	var (
		dec []byte
		err error
	)
	if dec, err = base64.StdEncoding.DecodeString(payload); err != nil {
		if dec, err = base64.URLEncoding.DecodeString(payload); err != nil {
			dec, err = base64.RawURLEncoding.DecodeString(payload)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("content_base64: decode failed (use standard or URL-safe base64): %w", err)
	}
	if len(dec) > maxDecodedBytes {
		return nil, fmt.Errorf("decoded file exceeds maximum size (%d bytes; max %d)", len(dec), maxDecodedBytes)
	}
	return dec, nil
}

// mimeTypeFromFileExtension returns a MIME type for well-known extensions, or "" if unknown.
func mimeTypeFromFileExtension(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".doc":
		return "application/msword"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".zip":
		return "application/zip"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".txt":
		return "text/plain"
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return ""
	}
}

// resolveDriveUploadMimeType returns explicit MIME when set, otherwise from the file name extension.
func resolveDriveUploadMimeType(explicitMime, fileName string) string {
	if strings.TrimSpace(explicitMime) != "" {
		return strings.TrimSpace(explicitMime)
	}
	return mimeTypeFromFileExtension(fileName)
}

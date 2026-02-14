package validate

import (
	"fmt"
	"regexp"
)

// driveIDRE matches valid Google Drive file/folder IDs.
// Drive IDs are alphanumeric with hyphens and underscores, typically 25-60 chars,
// plus the special "root" literal.
var driveIDRE = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)

// DriveID validates that the given string is a safe Google Drive resource ID.
// This prevents query injection when IDs are interpolated into Drive API queries.
func DriveID(id string) error {
	if !driveIDRE.MatchString(id) {
		return fmt.Errorf("invalid Drive resource ID %q — expected alphanumeric characters, hyphens, and underscores", id)
	}
	return nil
}

// emailRE matches basic email format: local@domain with at least one dot in domain.
var emailRE = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email validates that the given string looks like a valid email address.
func Email(email string) error {
	if len(email) > 254 {
		return fmt.Errorf("email address too long (max 254 characters)")
	}
	if !emailRE.MatchString(email) {
		return fmt.Errorf("invalid email address %q", email)
	}
	return nil
}

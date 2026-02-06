package middleware

import (
	"errors"
	"fmt"

	"google.golang.org/api/googleapi"
)

// HandleGoogleAPIError translates Google API errors into agent-actionable messages.
// These messages tell the AI what to do next, not the end user.
func HandleGoogleAPIError(err error) error {
	if err == nil {
		return nil
	}

	var googleErr *googleapi.Error
	if errors.As(err, &googleErr) {
		switch googleErr.Code {
		case 400:
			return fmt.Errorf(
				"bad request — check that all required parameters are provided and valid. Detail: %s",
				googleErr.Message)
		case 401:
			return fmt.Errorf(
				"authentication expired for this user — call start_google_auth tool to re-authenticate, " +
					"or verify the OAuth configuration is correct")
		case 403:
			return fmt.Errorf(
				"permission denied — the required OAuth scope may not be granted. "+
					"Suggest the user re-authenticate with broader scopes. Detail: %s", googleErr.Message)
		case 404:
			return fmt.Errorf(
				"resource not found — verify the ID is correct and the user has access to it")
		case 409:
			return fmt.Errorf(
				"conflict — the resource was modified by another process. Retry with the latest version. Detail: %s",
				googleErr.Message)
		case 429:
			return fmt.Errorf(
				"rate limit exceeded for this Google API — wait 30-60 seconds before retrying this tool call")
		case 500, 502, 503:
			return fmt.Errorf(
				"Google API server error (%d) — this is a transient issue, retry after a few seconds. Detail: %s",
				googleErr.Code, googleErr.Message)
		default:
			return fmt.Errorf("Google API error (%d): %s", googleErr.Code, googleErr.Message)
		}
	}

	return err
}

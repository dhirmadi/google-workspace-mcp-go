package middleware

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestHandleGoogleAPIError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantNil     bool
		wantContain string
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:        "400 bad request",
			err:         &googleapi.Error{Code: 400, Message: "invalid field"},
			wantContain: "bad request",
		},
		{
			name:        "401 auth expired",
			err:         &googleapi.Error{Code: 401, Message: "token expired"},
			wantContain: "start_google_auth",
		},
		{
			name:        "403 permission denied generic",
			err:         &googleapi.Error{Code: 403, Message: "insufficient scope"},
			wantContain: "permission denied",
		},
		{
			name:        "403 sharing outside org",
			err:         &googleapi.Error{Code: 403, Message: "Sharing outside of the organization is not allowed"},
			wantContain: "Workspace policy",
		},
		{
			name:        "403 not allowed to share",
			err:         &googleapi.Error{Code: 403, Message: "User is not allowed to share this file"},
			wantContain: "Workspace policy",
		},
		{
			name:        "404 not found",
			err:         &googleapi.Error{Code: 404, Message: "file not found"},
			wantContain: "not found",
		},
		{
			name:        "409 conflict",
			err:         &googleapi.Error{Code: 409, Message: "version mismatch"},
			wantContain: "conflict",
		},
		{
			name:        "429 rate limit",
			err:         &googleapi.Error{Code: 429, Message: "quota exceeded"},
			wantContain: "rate limit",
		},
		{
			name:        "500 server error",
			err:         &googleapi.Error{Code: 500, Message: "internal"},
			wantContain: "server error",
		},
		{
			name:        "502 server error",
			err:         &googleapi.Error{Code: 502, Message: "bad gateway"},
			wantContain: "server error",
		},
		{
			name:        "503 server error",
			err:         &googleapi.Error{Code: 503, Message: "unavailable"},
			wantContain: "server error",
		},
		{
			name:        "unknown google error code",
			err:         &googleapi.Error{Code: 418, Message: "teapot"},
			wantContain: "Google API error (418)",
		},
		{
			name:        "non-google error passed through",
			err:         fmt.Errorf("connection refused"),
			wantContain: "connection refused",
		},
		{
			name:        "wrapped google error",
			err:         fmt.Errorf("doing thing: %w", &googleapi.Error{Code: 404, Message: "gone"}),
			wantContain: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleGoogleAPIError(tt.err)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil error")
			}
			if !strings.Contains(got.Error(), tt.wantContain) {
				t.Errorf("error %q should contain %q", got.Error(), tt.wantContain)
			}
		})
	}
}

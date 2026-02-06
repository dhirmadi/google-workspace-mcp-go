package auth

// BaseScopes are always required for user identity.
var BaseScopes = []string{
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
	"openid",
}

// ServiceScopes maps service names to their full-access OAuth scopes.
// Broader scopes already imply narrower ones — no redundant requests.
var ServiceScopes = map[string][]string{
	"gmail": {
		"https://www.googleapis.com/auth/gmail.modify",
		"https://www.googleapis.com/auth/gmail.send",
		"https://www.googleapis.com/auth/gmail.labels",
		"https://www.googleapis.com/auth/gmail.settings.basic",
	},
	"drive": {
		"https://www.googleapis.com/auth/drive",
	},
	"calendar": {
		"https://www.googleapis.com/auth/calendar",
	},
	"docs": {
		"https://www.googleapis.com/auth/documents",
	},
	"sheets": {
		"https://www.googleapis.com/auth/spreadsheets",
	},
	"chat": {
		"https://www.googleapis.com/auth/chat.messages",
		"https://www.googleapis.com/auth/chat.spaces",
	},
	"forms": {
		"https://www.googleapis.com/auth/forms.body",
		"https://www.googleapis.com/auth/forms.responses.readonly",
	},
	"slides": {
		"https://www.googleapis.com/auth/presentations",
	},
	"tasks": {
		"https://www.googleapis.com/auth/tasks",
	},
	"contacts": {
		"https://www.googleapis.com/auth/contacts",
	},
	"search": {
		"https://www.googleapis.com/auth/cse",
	},
	"appscript": {
		"https://www.googleapis.com/auth/script.projects",
		"https://www.googleapis.com/auth/script.deployments",
		"https://www.googleapis.com/auth/script.processes",
		"https://www.googleapis.com/auth/script.metrics",
		"https://www.googleapis.com/auth/drive.file",
	},
}

// ReadOnlyScopes maps service names to their read-only OAuth scopes.
// Used when --read-only is set.
var ReadOnlyScopes = map[string][]string{
	"gmail": {
		"https://www.googleapis.com/auth/gmail.readonly",
	},
	"drive": {
		"https://www.googleapis.com/auth/drive.readonly",
	},
	"calendar": {
		"https://www.googleapis.com/auth/calendar.readonly",
	},
	"docs": {
		"https://www.googleapis.com/auth/documents.readonly",
	},
	"sheets": {
		"https://www.googleapis.com/auth/spreadsheets.readonly",
	},
	"chat": {
		"https://www.googleapis.com/auth/chat.messages.readonly",
		"https://www.googleapis.com/auth/chat.spaces.readonly",
	},
	"forms": {
		"https://www.googleapis.com/auth/forms.body.readonly",
		"https://www.googleapis.com/auth/forms.responses.readonly",
	},
	"slides": {
		"https://www.googleapis.com/auth/presentations.readonly",
	},
	"tasks": {
		"https://www.googleapis.com/auth/tasks.readonly",
	},
	"contacts": {
		"https://www.googleapis.com/auth/contacts.readonly",
	},
	"search": {
		"https://www.googleapis.com/auth/cse",
	},
	"appscript": {
		"https://www.googleapis.com/auth/script.projects.readonly",
		"https://www.googleapis.com/auth/script.deployments.readonly",
		"https://www.googleapis.com/auth/script.processes",
		"https://www.googleapis.com/auth/script.metrics",
		"https://www.googleapis.com/auth/drive.readonly",
	},
}

// AllScopes returns the combined set of scopes for the given services and mode.
func AllScopes(services []string, readOnly bool) []string {
	seen := make(map[string]bool)
	var scopes []string

	for _, s := range BaseScopes {
		if !seen[s] {
			scopes = append(scopes, s)
			seen[s] = true
		}
	}

	scopeMap := ServiceScopes
	if readOnly {
		scopeMap = ReadOnlyScopes
	}

	// If no services specified, include all
	if len(services) == 0 {
		for _, svcScopes := range scopeMap {
			for _, s := range svcScopes {
				if !seen[s] {
					scopes = append(scopes, s)
					seen[s] = true
				}
			}
		}
	} else {
		for _, svc := range services {
			for _, s := range scopeMap[svc] {
				if !seen[s] {
					scopes = append(scopes, s)
					seen[s] = true
				}
			}
		}
	}

	return scopes
}

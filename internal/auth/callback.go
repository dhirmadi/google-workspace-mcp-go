package auth

import (
	"fmt"
	"log/slog"
	"net/http"
)

// ClientInvalidator is called after successful OAuth to clear cached API clients.
type ClientInvalidator interface {
	InvalidateClient(userEmail string)
}

// OAuthCallbackHandler returns an http.HandlerFunc that handles the OAuth 2.0 callback.
// It exchanges the authorization code for a token and persists it.
// The state parameter carries the user's email address.
// If invalidator is non-nil, cached API clients are evicted on successful auth
// so the next call picks up the fresh token.
func OAuthCallbackHandler(oauthMgr *OAuthManager, invalidator ClientInvalidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state") // user email
		errMsg := r.URL.Query().Get("error")

		if errMsg != "" {
			slog.Error("OAuth callback error", "error", errMsg, "state", state)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, renderErrorPage(errMsg))
			return
		}

		if code == "" {
			slog.Error("OAuth callback missing code", "state", state)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, renderErrorPage("No authorization code received from Google."))
			return
		}

		if state == "" {
			slog.Error("OAuth callback missing state (user email)")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, renderErrorPage("Missing user email in OAuth state. Please restart the authentication from the MCP client."))
			return
		}

		// Exchange code for token and persist it
		_, err := oauthMgr.ExchangeCode(r.Context(), code, state)
		if err != nil {
			slog.Error("OAuth token exchange failed", "email", state, "error", err)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, renderErrorPage(fmt.Sprintf("Token exchange failed: %v", err)))
			return
		}

		// Evict any cached HTTP client so the next API call rebuilds from
		// the freshly persisted token instead of reusing a stale one.
		if invalidator != nil {
			invalidator.InvalidateClient(state)
			slog.Info("invalidated cached client after re-auth", "email", state)
		}

		slog.Info("OAuth authentication successful", "email", state)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, renderSuccessPage(state))
	}
}

func renderSuccessPage(email string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Authentication Successful</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: 'Red Hat Display', 'Segoe UI', system-ui, -apple-system, sans-serif;
      background: linear-gradient(135deg, #1a1a1a 0%%, #2d2d2d 100%%);
      color: #e0e0e0;
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .card {
      background: #2d2d2d;
      border: 1px solid #444;
      border-radius: 16px;
      padding: 48px;
      max-width: 480px;
      width: 90%%;
      text-align: center;
      box-shadow: 0 24px 48px rgba(0,0,0,0.4);
    }
    .icon {
      width: 64px;
      height: 64px;
      background: #4caf50;
      border-radius: 50%%;
      display: flex;
      align-items: center;
      justify-content: center;
      margin: 0 auto 24px;
    }
    .icon svg { width: 32px; height: 32px; fill: white; }
    h1 {
      font-size: 24px;
      font-weight: 600;
      color: #ffffff;
      margin-bottom: 12px;
    }
    .email {
      font-size: 16px;
      color: #ee0000;
      font-weight: 500;
      margin-bottom: 8px;
    }
    .message {
      font-size: 14px;
      color: #aaa;
      line-height: 1.6;
      margin-bottom: 32px;
    }
    .badge {
      display: inline-block;
      background: rgba(238, 0, 0, 0.1);
      border: 1px solid rgba(238, 0, 0, 0.3);
      color: #ee0000;
      padding: 6px 16px;
      border-radius: 20px;
      font-size: 12px;
      font-weight: 500;
      letter-spacing: 0.5px;
      text-transform: uppercase;
    }
    .close-hint {
      margin-top: 24px;
      font-size: 13px;
      color: #666;
    }
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">
      <svg viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
    </div>
    <h1>Authentication Successful</h1>
    <p class="email">%s</p>
    <p class="message">
      Your Google Workspace account has been connected.<br>
      All MCP tools are now available for this account.
    </p>
    <span class="badge">Google Workspace MCP</span>
    <p class="close-hint">You can close this window.</p>
  </div>
</body>
</html>`, email)
}

func renderErrorPage(errMsg string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Authentication Failed</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: 'Red Hat Display', 'Segoe UI', system-ui, -apple-system, sans-serif;
      background: linear-gradient(135deg, #1a1a1a 0%%, #2d2d2d 100%%);
      color: #e0e0e0;
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .card {
      background: #2d2d2d;
      border: 1px solid #444;
      border-radius: 16px;
      padding: 48px;
      max-width: 480px;
      width: 90%%;
      text-align: center;
      box-shadow: 0 24px 48px rgba(0,0,0,0.4);
    }
    .icon {
      width: 64px;
      height: 64px;
      background: #ee0000;
      border-radius: 50%%;
      display: flex;
      align-items: center;
      justify-content: center;
      margin: 0 auto 24px;
    }
    .icon svg { width: 32px; height: 32px; fill: white; }
    h1 {
      font-size: 24px;
      font-weight: 600;
      color: #ffffff;
      margin-bottom: 16px;
    }
    .error-msg {
      font-size: 14px;
      color: #ff6b6b;
      background: rgba(238, 0, 0, 0.1);
      border: 1px solid rgba(238, 0, 0, 0.2);
      border-radius: 8px;
      padding: 16px;
      margin-bottom: 24px;
      line-height: 1.5;
      word-break: break-word;
    }
    .message {
      font-size: 14px;
      color: #aaa;
      line-height: 1.6;
    }
    .badge {
      display: inline-block;
      background: rgba(238, 0, 0, 0.1);
      border: 1px solid rgba(238, 0, 0, 0.3);
      color: #ee0000;
      padding: 6px 16px;
      border-radius: 20px;
      font-size: 12px;
      font-weight: 500;
      letter-spacing: 0.5px;
      text-transform: uppercase;
      margin-top: 24px;
    }
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">
      <svg viewBox="0 0 24 24"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
    </div>
    <h1>Authentication Failed</h1>
    <div class="error-msg">%s</div>
    <p class="message">Please return to the MCP client and try again.</p>
    <span class="badge">Google Workspace MCP</span>
  </div>
</body>
</html>`, errMsg)
}

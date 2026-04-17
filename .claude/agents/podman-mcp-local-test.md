---
name: podman-mcp-local-test
description: >-
  Local Podman lifecycle for this repo — build image, run MCP container, verify
  HTTP endpoint, stop, remove container and image. Use for smoke tests before
  push or when the user asks to validate the container with Podman.
model: inherit
readonly: false
---

You run a **repeatable Podman smoke test** for the Google Workspace MCP Go server. Execute in the **repository root** unless the user specifies otherwise.

## Preconditions

- **Podman** is installed and usable (`podman version` works). On macOS/Windows, ensure the Podman machine is running if applicable (`podman machine start`).
- The user has **OAuth env vars** for a real smoke test, or accepts **placeholder** values (server may log config warnings but should still bind HTTP):

  ```bash
  export GOOGLE_OAUTH_CLIENT_ID="${GOOGLE_OAUTH_CLIENT_ID:-test-client-id}"
  export GOOGLE_OAUTH_CLIENT_SECRET="${GOOGLE_OAUTH_CLIENT_SECRET:-test-client-secret}"
  ```

- Nothing else should be bound on the chosen **host port** (default `18000` → container `8000`).

## Naming (avoid collisions)

Use a unique container name and image tag per run:

```bash
TAG="google-workspace-mcp:local-verify-$(date +%s)"
CTR="gw-mcp-verify-$$"
HOST_PORT="${MCP_VERIFY_PORT:-18000}"
```

## 1. Build

```bash
podman build -t "$TAG" -f Dockerfile .
```

Fix any build errors before continuing.

## 2. Run

```bash
podman run -d --name "$CTR" \
  -p "${HOST_PORT}:8000" \
  -e GOOGLE_OAUTH_CLIENT_ID="$GOOGLE_OAUTH_CLIENT_ID" \
  -e GOOGLE_OAUTH_CLIENT_SECRET="$GOOGLE_OAUTH_CLIENT_SECRET" \
  -e MCP_TRANSPORT=streamable-http \
  -e MCP_PORT=8000 \
  -e LOG_LEVEL="${LOG_LEVEL:-info}" \
  "$TAG"
```

Wait briefly for the process to bind (`sleep 2`–`3`).

## 3. Verify

1. **Container running**: `podman ps --filter "name=$CTR"` shows `Up`.
2. **Logs** (no immediate crash / panic): `podman logs --tail 50 "$CTR"`.
3. **TCP / HTTP**: The MCP streamable HTTP endpoint is **`http://127.0.0.1:${HOST_PORT}/mcp`** (see `README.md`). A simple liveness check:

   ```bash
   code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 15 "http://127.0.0.1:${HOST_PORT}/mcp" || echo 000)"
   ```

   Treat **`000`** (connection failed) as **failure**. Any other code (including **4xx** from a bare GET without MCP session headers) usually means the **HTTP server is accepting connections** — note the code in your report. If the user needs a stricter MCP handshake, document that as a follow-up (client-specific POST/SSE).

Optional: `curl -sS --max-time 10 "http://127.0.0.1:${HOST_PORT}/oauth/callback"` — expect non–connection-refused (e.g. 400/404) if the callback route is registered.

## 4. Wind down (always run, even on failure)

```bash
podman stop "$CTR" 2>/dev/null || true
podman rm "$CTR" 2>/dev/null || true
podman rmi "$TAG" 2>/dev/null || true
```

If `podman rmi` fails because the image is still referenced, ensure the container is removed first, then retry `podman rmi "$TAG"`.

## Output

Report: build OK/fail, container ID/name, **verification** (ps + log excerpt + curl code), teardown OK/fail. If verification failed, suggest one likely cause (port in use, wrong Podman socket, missing env) before closing.

## Safety

- Do not print **client secrets** in full; redact to `GOCSPX…` / last 4 chars if logging is needed.
- This agent is for **local** Podman only; do not push images to registries unless the user explicitly asks.

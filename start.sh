#!/usr/bin/env bash
# start.sh — Build and run the Google Workspace MCP server in Docker Desktop
#
# Usage:
#   ./start.sh <CLIENT_ID> <CLIENT_SECRET> [OPTIONS]
#
# Examples:
#   ./start.sh "123456.apps.googleusercontent.com" "GOCSPX-abc123"
#   ./start.sh "123456.apps.googleusercontent.com" "GOCSPX-abc123" --port 9000
#   ./start.sh "123456.apps.googleusercontent.com" "GOCSPX-abc123" --services gmail,drive,calendar --email user@gmail.com
#
# Options:
#   --port PORT         HTTP port (default: 8000)
#   --services SVCS     Comma-separated services to enable (default: all)
#                       Options: gmail,drive,calendar,docs,sheets,chat,
#                                forms,slides,tasks,contacts,search,appscript
#   --email EMAIL       Default user email for authentication
#   --cse-id ID         Google Custom Search Engine ID (for search tools)
#   --log-level LEVEL   Log level: debug, info, warn, error (default: info)
#   --rebuild           Force rebuild of the Docker image
#   --stop              Stop and remove existing container, then exit

set -euo pipefail

# ── Defaults ──────────────────────────────────────────────────────
CONTAINER_NAME="google-workspace-mcp"
IMAGE_NAME="google-workspace-mcp:latest"
PORT=8000
SERVICES=""
EMAIL=""
CSE_ID=""
LOG_LEVEL="info"
REBUILD=false
STOP_ONLY=false

# ── Color output ──────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

info()  { echo -e "${CYAN}[INFO]${NC}  $*"; }
ok()    { echo -e "${GREEN}[OK]${NC}    $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

# ── Parse arguments ───────────────────────────────────────────────
if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <GOOGLE_OAUTH_CLIENT_ID> <GOOGLE_OAUTH_CLIENT_SECRET> [OPTIONS]"
    echo ""
    echo "Run '$0 --help' for full options."
    exit 1
fi

# Handle --stop and --help before positional args
case "${1:-}" in
    --stop)
        STOP_ONLY=true
        ;;
    --help|-h)
        head -20 "$0" | grep '^#' | sed 's/^# \?//'
        exit 0
        ;;
esac

if [[ "$STOP_ONLY" == "true" ]]; then
    info "Stopping container ${CONTAINER_NAME}..."
    docker stop "$CONTAINER_NAME" 2>/dev/null && ok "Container stopped." || warn "Container was not running."
    docker rm "$CONTAINER_NAME" 2>/dev/null && ok "Container removed." || true
    exit 0
fi

# Positional arguments
CLIENT_ID="${1:?Error: GOOGLE_OAUTH_CLIENT_ID is required as first argument}"
CLIENT_SECRET="${2:?Error: GOOGLE_OAUTH_CLIENT_SECRET is required as second argument}"
shift 2

# Parse optional flags
while [[ $# -gt 0 ]]; do
    case "$1" in
        --port)       PORT="$2";       shift 2 ;;
        --services)   SERVICES="$2";   shift 2 ;;
        --email)      EMAIL="$2";      shift 2 ;;
        --cse-id)     CSE_ID="$2";     shift 2 ;;
        --log-level)  LOG_LEVEL="$2";  shift 2 ;;
        --rebuild)    REBUILD=true;    shift   ;;
        *)
            error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# ── Validate ──────────────────────────────────────────────────────
if [[ -z "$CLIENT_ID" || -z "$CLIENT_SECRET" ]]; then
    error "Both GOOGLE_OAUTH_CLIENT_ID and GOOGLE_OAUTH_CLIENT_SECRET are required."
    exit 1
fi

if ! command -v docker &>/dev/null; then
    error "Docker is not installed or not in PATH."
    exit 1
fi

if ! docker info &>/dev/null 2>&1; then
    error "Docker daemon is not running. Start Docker Desktop first."
    exit 1
fi

# ── Stop existing container if running ────────────────────────────
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    warn "Existing container '${CONTAINER_NAME}' found — removing..."
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm "$CONTAINER_NAME" 2>/dev/null || true
    ok "Old container removed."
fi

# ── Build image ───────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "$REBUILD" == "true" ]] || ! docker image inspect "$IMAGE_NAME" &>/dev/null; then
    info "Building Docker image '${IMAGE_NAME}'..."
    docker build -t "$IMAGE_NAME" "$SCRIPT_DIR"
    ok "Image built successfully."
else
    ok "Image '${IMAGE_NAME}' already exists (use --rebuild to force)."
fi

# ── Build docker run command ──────────────────────────────────────
RUN_ARGS=(
    docker run -d
    --name "$CONTAINER_NAME"
    -p "${PORT}:8000"
    -e "GOOGLE_OAUTH_CLIENT_ID=${CLIENT_ID}"
    -e "GOOGLE_OAUTH_CLIENT_SECRET=${CLIENT_SECRET}"
    -e "MCP_TRANSPORT=streamable-http"
    -e "MCP_PORT=8000"
    -e "WORKSPACE_MCP_HOST=0.0.0.0"
    -e "WORKSPACE_MCP_BASE_URI=http://localhost:${PORT}"
    -e "WORKSPACE_MCP_CREDENTIALS_DIR=/data/credentials"
    -e "LOG_LEVEL=${LOG_LEVEL}"
    -v "mcp-credentials:/data/credentials"
    --restart unless-stopped
)

if [[ -n "$EMAIL" ]]; then
    RUN_ARGS+=(-e "USER_GOOGLE_EMAIL=${EMAIL}")
    RUN_ARGS+=(-e "MCP_SINGLE_USER_MODE=true")
fi

if [[ -n "$SERVICES" ]]; then
    RUN_ARGS+=(-e "ENABLED_SERVICES=${SERVICES}")
fi

if [[ -n "$CSE_ID" ]]; then
    RUN_ARGS+=(-e "GOOGLE_CSE_ID=${CSE_ID}")
fi

RUN_ARGS+=("$IMAGE_NAME")

# ── Start container ───────────────────────────────────────────────
info "Starting MCP server container..."
CONTAINER_ID=$("${RUN_ARGS[@]}")
ok "Container started: ${CONTAINER_ID:0:12}"

# ── Wait for healthy startup ─────────────────────────────────────
info "Waiting for server to start..."
sleep 2

if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    ok "Server is running!"
else
    error "Container failed to start. Logs:"
    docker logs "$CONTAINER_NAME" 2>&1 | tail -20
    exit 1
fi

# ── Print summary ─────────────────────────────────────────────────
echo ""
echo -e "${GREEN}══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  Google Workspace MCP Server is running${NC}"
echo -e "${GREEN}══════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "  Endpoint:    ${CYAN}http://localhost:${PORT}/mcp${NC}"
if [[ -n "$SERVICES" ]]; then
echo -e "  Services:    ${CYAN}${SERVICES}${NC}"
else
echo -e "  Services:    all (gmail,drive,calendar,docs,sheets,chat,forms,slides,tasks,contacts,search,appscript)"
fi
echo -e "  Log level:   ${LOG_LEVEL}"
echo -e "  Container:   ${CONTAINER_NAME}"
if [[ -n "$EMAIL" ]]; then
echo -e "  User email:  ${EMAIL} (single-user mode)"
fi
echo ""
echo -e "  ${YELLOW}Next steps:${NC}"
echo -e "  1. Authenticate: The AI agent calls start_google_auth"
echo -e "     with the user's email to get an OAuth consent URL."
echo -e "  2. The user opens the URL in a browser and grants access."
echo -e "  3. All 136 tools are then available via the MCP endpoint."
echo ""
echo -e "  ${CYAN}Useful commands:${NC}"
echo -e "    docker logs -f ${CONTAINER_NAME}     # Stream logs"
echo -e "    docker stop ${CONTAINER_NAME}         # Stop server"
echo -e "    ./start.sh --stop                     # Stop & remove"
echo ""

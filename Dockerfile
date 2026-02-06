# syntax=docker/dockerfile:1

# ── Build stage ──────────────────────────────────────────────────
FROM golang:1.24 AS builder

WORKDIR /src

# Cache dependency downloads
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source and build
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /bin/server ./cmd/server

# ── Runtime stage ────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/server /server
COPY configs/ /configs/

# Default environment
ENV MCP_TRANSPORT=streamable-http
ENV MCP_PORT=8000

EXPOSE 8000

ENTRYPOINT ["/server"]

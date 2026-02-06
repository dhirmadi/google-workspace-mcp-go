package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LoggingMiddleware returns MCP SDK middleware that logs incoming requests
// and outgoing responses using structured logging.
func LoggingMiddleware(logger *slog.Logger) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			logger.InfoContext(ctx, "handling request", "method", method)

			result, err := next(ctx, method, req)

			duration := time.Since(start)
			if err != nil {
				logger.ErrorContext(ctx, "request failed",
					"method", method,
					"duration", duration,
					"error", err,
				)
			} else {
				logger.InfoContext(ctx, "request completed",
					"method", method,
					"duration", duration,
				)
			}

			return result, err
		}
	}
}

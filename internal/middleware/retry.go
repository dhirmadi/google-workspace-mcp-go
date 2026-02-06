package middleware

import (
	"context"
	"errors"
	"time"

	"google.golang.org/api/googleapi"
)

// DefaultMaxRetries is the default maximum number of retry attempts for 429 errors.
const DefaultMaxRetries = 3

// WithRetry executes fn with exponential backoff on 429 (rate limit) errors.
// Only retries on 429 errors; all other errors are returned immediately.
func WithRetry(ctx context.Context, maxAttempts int, fn func() error) error {
	if maxAttempts <= 0 {
		maxAttempts = DefaultMaxRetries
	}

	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		// Only retry on 429 rate limit errors
		var googleErr *googleapi.Error
		if !errors.As(err, &googleErr) || googleErr.Code != 429 {
			return err
		}

		// Exponential backoff: 1s, 2s, 4s, 8s, ...
		backoff := time.Duration(1<<uint(attempt)) * time.Second
		select {
		case <-time.After(backoff):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

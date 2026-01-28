package processor

import (
	"context"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
)

// Counter defines the interface for counting entities in the system.
// Implementations are responsible for providing count operations for specific entity types.
type Counter interface {
	Count(ctx context.Context) (int, error)
}

// APIInternal implements the APIInternalProcessor interface for internal statistics operations.
// It aggregates counts from multiple sources to provide comprehensive service statistics.
type APIInternal struct {
	user Counter
	url  Counter
}

// NewAPIInternal creates a new APIInternal statistics processor.
//
// Parameters:
//   - user: Counter implementation for user entities
//   - url: Counter implementation for URL entities
//
// Returns:
//   - *APIInternal: Configured statistics processor instance
func NewAPIInternal(user Counter, url Counter) *APIInternal {
	return &APIInternal{
		user: user,
		url:  url,
	}
}

// Process retrieves and aggregates statistics from all configured counters.
// It collects user count and URL count, returning them in a unified StatsResponse.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//
// Returns:
//   - model.StatsResponse: Structure containing URLsCount and UsersCount fields
//   - error: Returns an error if any counter operation fails
//
// The method returns an empty StatsResponse and an error if either count operation fails.
// Error messages indicate which specific counter failed (users or URLs).
func (a *APIInternal) Process(ctx context.Context) (model.StatsResponse, error) {
	usersCount, err := a.user.Count(ctx)
	if err != nil {
		return model.StatsResponse{}, fmt.Errorf("count user: %w", err)
	}
	urlsCount, err := a.url.Count(ctx)
	if err != nil {
		return model.StatsResponse{}, fmt.Errorf("count urls: %w", err)
	}
	return model.StatsResponse{
		URLsCount:  urlsCount,
		UsersCount: usersCount,
	}, nil
}

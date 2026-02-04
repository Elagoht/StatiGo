// Package services provides example service patterns for Statigo applications.
package services

import (
	"context"
	"log/slog"
	"time"

	"statigo/framework/client"
)

// ExampleService demonstrates the service pattern for external API calls.
type ExampleService struct {
	client *client.Client
	logger *slog.Logger
}

// NewExampleService creates a new example service.
func NewExampleService(client *client.Client, logger *slog.Logger) *ExampleService {
	return &ExampleService{
		client: client,
		logger: logger,
	}
}

// ExampleData represents data returned by the example service.
type ExampleData struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// GetData fetches example data from an external API.
// This demonstrates the pattern for making HTTP requests with the framework client.
func (s *ExampleService) GetData(ctx context.Context, id string) (*ExampleData, error) {
	s.logger.Info("fetching example data", slog.String("id", id))

	// Example of how to use the HTTP client
	// In a real application, you would make actual API calls here:
	//
	// var data ExampleData
	// err := s.client.Get(ctx, "/api/data/"+id, &data)
	// if err != nil {
	//     return nil, err
	// }
	// return &data, nil

	// For demonstration, return mock data
	return &ExampleData{
		ID:        id,
		Title:     "Example Title",
		Content:   "This is example content from the service.",
		CreatedAt: time.Now(),
	}, nil
}

// ListData fetches a list of example data.
func (s *ExampleService) ListData(ctx context.Context, limit int) ([]ExampleData, error) {
	s.logger.Info("listing example data", slog.Int("limit", limit))

	// Example pattern for list operations
	items := make([]ExampleData, 0, limit)
	for i := 0; i < limit; i++ {
		items = append(items, ExampleData{
			ID:        string(rune('A' + i)),
			Title:     "Item " + string(rune('A'+i)),
			Content:   "Content for item " + string(rune('A'+i)),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		})
	}

	return items, nil
}

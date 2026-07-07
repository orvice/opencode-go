package opencode

import (
	"context"
	"net/http"
)

// Config is the opencode configuration. The schema is large and evolves with
// the server, so it is kept loosely typed; see the server's /doc endpoint
// for the full schema.
type Config map[string]any

// ConfigService accesses the /config endpoints.
type ConfigService struct {
	client *Client
}

// Get returns the current configuration.
func (s *ConfigService) Get(ctx context.Context) (Config, error) {
	var out Config
	err := s.client.do(ctx, http.MethodGet, "/config", nil, nil, &out)
	return out, err
}

// Update applies a partial configuration update and returns the new config.
func (s *ConfigService) Update(ctx context.Context, patch Config) (Config, error) {
	var out Config
	err := s.client.do(ctx, http.MethodPatch, "/config", nil, patch, &out)
	return out, err
}

// ConfigProviders lists configured providers and their default models.
type ConfigProviders struct {
	Providers []Provider `json:"providers"`
	// Default maps provider ID to its default model ID.
	Default map[string]string `json:"default"`
}

// Providers lists providers and default models from the configuration.
func (s *ConfigService) Providers(ctx context.Context) (*ConfigProviders, error) {
	var out ConfigProviders
	err := s.client.do(ctx, http.MethodGet, "/config/providers", nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

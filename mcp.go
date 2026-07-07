package opencode

import (
	"context"
	"net/http"
	"net/url"
)

// MCPService accesses the /mcp endpoints.
type MCPService struct {
	client *Client
}

// MCPStatus reports the state of an MCP server. Status is "connected",
// "disabled", "failed", "needs_auth" or "needs_client_registration";
// Error is set for failed states.
type MCPStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Status returns the status of all MCP servers, keyed by name.
func (s *MCPService) Status(ctx context.Context) (map[string]MCPStatus, error) {
	var out map[string]MCPStatus
	err := s.client.do(ctx, http.MethodGet, "/mcp", nil, nil, &out)
	return out, err
}

// MCPConfig configures an MCP server. Type is "local" (Command/Environment)
// or "remote" (URL/Headers/OAuth).
type MCPConfig struct {
	Type string `json:"type"`
	// Local server fields.
	Command     []string          `json:"command,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	// Remote server fields.
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	// OAuth is an OAuth config object, or false to disable auto-detection.
	OAuth any `json:"oauth,omitempty"`
	// Common fields.
	Enabled *bool `json:"enabled,omitempty"`
	Timeout int   `json:"timeout,omitempty"`
}

// Add dynamically adds an MCP server and returns the updated status map.
func (s *MCPService) Add(ctx context.Context, name string, config MCPConfig) (map[string]MCPStatus, error) {
	body := map[string]any{"name": name, "config": config}
	var out map[string]MCPStatus
	err := s.client.do(ctx, http.MethodPost, "/mcp", nil, body, &out)
	return out, err
}

// Connect connects a configured MCP server.
func (s *MCPService) Connect(ctx context.Context, name string) error {
	return s.client.do(ctx, http.MethodPost, "/mcp/"+name+"/connect", nil, nil, nil)
}

// Disconnect disconnects a configured MCP server.
func (s *MCPService) Disconnect(ctx context.Context, name string) error {
	return s.client.do(ctx, http.MethodPost, "/mcp/"+name+"/disconnect", nil, nil, nil)
}

// LSPService accesses the /lsp endpoint.
type LSPService struct {
	client *Client
}

// Status returns the status of all language servers.
func (s *LSPService) Status(ctx context.Context) ([]LSPStatus, error) {
	var out []LSPStatus
	err := s.client.do(ctx, http.MethodGet, "/lsp", nil, nil, &out)
	return out, err
}

// FormatterService accesses the /formatter endpoint.
type FormatterService struct {
	client *Client
}

// Status returns the status of all formatters.
func (s *FormatterService) Status(ctx context.Context) ([]FormatterStatus, error) {
	var out []FormatterStatus
	err := s.client.do(ctx, http.MethodGet, "/formatter", nil, nil, &out)
	return out, err
}

// ToolService accesses the experimental /experimental/tool endpoints.
type ToolService struct {
	client *Client
}

// IDs returns the identifiers of all registered tools.
func (s *ToolService) IDs(ctx context.Context) ([]string, error) {
	var out []string
	err := s.client.do(ctx, http.MethodGet, "/experimental/tool/ids", nil, nil, &out)
	return out, err
}

// ToolInfo describes a registered tool and its JSON schema parameters.
type ToolInfo struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// List returns tools with their JSON schemas, resolved for the given
// provider and model.
func (s *ToolService) List(ctx context.Context, provider, model string) ([]ToolInfo, error) {
	q := url.Values{"provider": {provider}, "model": {model}}
	var out []ToolInfo
	err := s.client.do(ctx, http.MethodGet, "/experimental/tool", q, nil, &out)
	return out, err
}

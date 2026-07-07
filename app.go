package opencode

import (
	"context"
	"net/http"
)

// AppService accesses app-level endpoints: logging, agents, commands,
// paths and VCS information.
type AppService struct {
	client *Client
}

// LogLevel is the level of a log entry.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogParams is a log entry to write to the server log.
type LogParams struct {
	Service string         `json:"service"`
	Level   LogLevel       `json:"level"`
	Message string         `json:"message"`
	Extra   map[string]any `json:"extra,omitempty"`
}

// Log writes an entry to the server log.
func (s *AppService) Log(ctx context.Context, params LogParams) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, "/log", nil, params, &out)
	return out, err
}

// Agents lists the available agents.
func (s *AppService) Agents(ctx context.Context) ([]Agent, error) {
	var out []Agent
	err := s.client.do(ctx, http.MethodGet, "/agent", nil, nil, &out)
	return out, err
}

// Commands lists the available slash commands.
func (s *AppService) Commands(ctx context.Context) ([]Command, error) {
	var out []Command
	err := s.client.do(ctx, http.MethodGet, "/command", nil, nil, &out)
	return out, err
}

// Path returns the server's important directories.
func (s *AppService) Path(ctx context.Context) (*Path, error) {
	var out Path
	err := s.client.do(ctx, http.MethodGet, "/path", nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// VCS returns version control information for the current project.
func (s *AppService) VCS(ctx context.Context) (*VcsInfo, error) {
	var out VcsInfo
	err := s.client.do(ctx, http.MethodGet, "/vcs", nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GlobalService accesses the /global endpoints.
type GlobalService struct {
	client *Client
}

// Health returns the server health and version.
func (s *GlobalService) Health(ctx context.Context) (*Health, error) {
	var out Health
	err := s.client.do(ctx, http.MethodGet, "/global/health", nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// InstanceService accesses the /instance endpoints.
type InstanceService struct {
	client *Client
}

// Dispose destroys the current server instance.
func (s *InstanceService) Dispose(ctx context.Context) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, "/instance/dispose", nil, nil, &out)
	return out, err
}

// ProjectService accesses the /project endpoints.
type ProjectService struct {
	client *Client
}

// List lists all projects known to the server.
func (s *ProjectService) List(ctx context.Context) ([]Project, error) {
	var out []Project
	err := s.client.do(ctx, http.MethodGet, "/project", nil, nil, &out)
	return out, err
}

// Current returns the current project.
func (s *ProjectService) Current(ctx context.Context) (*Project, error) {
	var out Project
	err := s.client.do(ctx, http.MethodGet, "/project/current", nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

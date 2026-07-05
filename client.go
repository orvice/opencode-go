// Package opencode provides a Go client for the opencode server HTTP API.
//
// See https://opencode.ai/docs/server/ for details about the server.
// Start a server with:
//
//	opencode serve --port 4096
//
// Then create a client:
//
//	client, err := opencode.NewClient()
//	sessions, err := client.Session.List(ctx, nil)
package opencode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	// DefaultBaseURL is the default address of a locally running opencode server.
	DefaultBaseURL = "http://127.0.0.1:4096"

	// defaultUsername is the HTTP basic auth username used by opencode
	// when OPENCODE_SERVER_PASSWORD is set.
	defaultUsername = "opencode"
)

// Client is an opencode server API client.
// Create one with NewClient and access the API through its services.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	username   string
	password   string
	directory  string
	workspace  string

	Global    *GlobalService
	Event     *EventService
	Config    *ConfigService
	Provider  *ProviderService
	Auth      *AuthService
	Session   *SessionService
	Find      *FindService
	File      *FileService
	App       *AppService
	Project   *ProjectService
	TUI       *TUIService
	MCP       *MCPService
	LSP       *LSPService
	Formatter *FormatterService
	Tool      *ToolService
	Instance  *InstanceService
}

// Option configures a Client.
type Option func(*Client) error

// WithBaseURL sets the server base URL (default http://127.0.0.1:4096).
func WithBaseURL(raw string) Option {
	return func(c *Client) error {
		u, err := url.Parse(raw)
		if err != nil {
			return fmt.Errorf("opencode: invalid base URL %q: %w", raw, err)
		}
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("opencode: base URL %q must include scheme and host", raw)
		}
		c.baseURL = u
		return nil
	}
}

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) error {
		if hc == nil {
			return fmt.Errorf("opencode: http client must not be nil")
		}
		c.httpClient = hc
		return nil
	}
}

// WithPassword sets the HTTP basic auth password.
// Defaults to the OPENCODE_SERVER_PASSWORD environment variable.
func WithPassword(password string) Option {
	return func(c *Client) error {
		c.password = password
		return nil
	}
}

// WithUsername sets the HTTP basic auth username.
// Defaults to the OPENCODE_SERVER_USERNAME environment variable, or "opencode".
func WithUsername(username string) Option {
	return func(c *Client) error {
		c.username = username
		return nil
	}
}

// WithDirectory sets a default `directory` query parameter sent on every
// request, selecting the project directory the server operates on.
func WithDirectory(dir string) Option {
	return func(c *Client) error {
		c.directory = dir
		return nil
	}
}

// WithWorkspace sets a default `workspace` query parameter sent on every request.
func WithWorkspace(id string) Option {
	return func(c *Client) error {
		c.workspace = id
		return nil
	}
}

// NewClient creates a new opencode API client.
func NewClient(opts ...Option) (*Client, error) {
	base, _ := url.Parse(DefaultBaseURL)
	c := &Client{
		baseURL:    base,
		httpClient: &http.Client{},
		username:   os.Getenv("OPENCODE_SERVER_USERNAME"),
		password:   os.Getenv("OPENCODE_SERVER_PASSWORD"),
	}
	if c.username == "" {
		c.username = defaultUsername
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.Global = &GlobalService{client: c}
	c.Event = &EventService{client: c}
	c.Config = &ConfigService{client: c}
	c.Provider = &ProviderService{client: c}
	c.Auth = &AuthService{client: c}
	c.Session = &SessionService{client: c}
	c.Find = &FindService{client: c}
	c.File = &FileService{client: c}
	c.App = &AppService{client: c}
	c.Project = &ProjectService{client: c}
	c.TUI = &TUIService{client: c}
	c.MCP = &MCPService{client: c}
	c.LSP = &LSPService{client: c}
	c.Formatter = &FormatterService{client: c}
	c.Tool = &ToolService{client: c}
	c.Instance = &InstanceService{client: c}
	return c, nil
}

// BaseURL returns the configured server base URL.
func (c *Client) BaseURL() string { return c.baseURL.String() }

// newRequest builds an *http.Request for the given API path. The path must
// begin with "/" and already be escaped where needed.
func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body any) (*http.Request, error) {
	u := *c.baseURL
	u.Path = strings.TrimSuffix(u.Path, "/") + path

	q := u.Query()
	if c.directory != "" {
		q.Set("directory", c.directory)
	}
	if c.workspace != "" {
		q.Set("workspace", c.workspace)
	}
	for k, vs := range query {
		for _, v := range vs {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()

	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("opencode: encode request body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	return req, nil
}

// do performs a JSON request and decodes the response into out (if non-nil).
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	req, err := c.newRequest(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return newError(method, path, resp)
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("opencode: decode %s %s response: %w", method, path, err)
	}
	return nil
}

package opencode

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// SessionService accesses the /session endpoints.
type SessionService struct {
	client *Client
}

func sessionPath(sessionID string, rest ...string) string {
	parts := append([]string{"/session", url.PathEscape(sessionID)}, rest...)
	return strings.Join(parts, "/")
}

// SessionListParams filters the session list.
type SessionListParams struct {
	// Scope limits results; the only supported value is "project".
	Scope string
	// Path filters sessions by directory path.
	Path string
	// Roots, when true, returns only root sessions (no children).
	Roots *bool
	// Start is a pagination cursor (unix milliseconds).
	Start int64
	// Search filters sessions by a search string.
	Search string
	// Limit caps the number of returned sessions.
	Limit int
}

func (p *SessionListParams) query() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.Scope != "" {
		q.Set("scope", p.Scope)
	}
	if p.Path != "" {
		q.Set("path", p.Path)
	}
	if p.Roots != nil {
		q.Set("roots", strconv.FormatBool(*p.Roots))
	}
	if p.Start > 0 {
		q.Set("start", strconv.FormatInt(p.Start, 10))
	}
	if p.Search != "" {
		q.Set("search", p.Search)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	return q
}

// List lists sessions. params may be nil.
func (s *SessionService) List(ctx context.Context, params *SessionListParams) ([]Session, error) {
	var out []Session
	err := s.client.do(ctx, http.MethodGet, "/session", params.query(), nil, &out)
	return out, err
}

// SessionCreateParams configures a new session. All fields are optional.
type SessionCreateParams struct {
	ParentID    string            `json:"parentID,omitempty"`
	Title       string            `json:"title,omitempty"`
	Agent       string            `json:"agent,omitempty"`
	Model       *SessionModel     `json:"model,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
	Permission  PermissionRuleset `json:"permission,omitempty"`
	WorkspaceID string            `json:"workspaceID,omitempty"`
}

// Create creates a new session. params may be nil.
func (s *SessionService) Create(ctx context.Context, params *SessionCreateParams) (*Session, error) {
	if params == nil {
		params = &SessionCreateParams{}
	}
	var out Session
	err := s.client.do(ctx, http.MethodPost, "/session", nil, params, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a session by ID.
func (s *SessionService) Get(ctx context.Context, sessionID string) (*Session, error) {
	var out Session
	err := s.client.do(ctx, http.MethodGet, sessionPath(sessionID), nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SessionUpdateTime updates session timestamps.
type SessionUpdateTime struct {
	Archived int64 `json:"archived,omitempty"`
}

// SessionUpdateParams updates session properties.
type SessionUpdateParams struct {
	Title      string             `json:"title,omitempty"`
	Metadata   map[string]any     `json:"metadata,omitempty"`
	Permission PermissionRuleset  `json:"permission,omitempty"`
	Time       *SessionUpdateTime `json:"time,omitempty"`
}

// Update updates session properties.
func (s *SessionService) Update(ctx context.Context, sessionID string, params *SessionUpdateParams) (*Session, error) {
	if params == nil {
		params = &SessionUpdateParams{}
	}
	var out Session
	err := s.client.do(ctx, http.MethodPatch, sessionPath(sessionID), nil, params, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes a session and all its data.
func (s *SessionService) Delete(ctx context.Context, sessionID string) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodDelete, sessionPath(sessionID), nil, nil, &out)
	return out, err
}

// Children lists the child sessions of a session.
func (s *SessionService) Children(ctx context.Context, sessionID string) ([]Session, error) {
	var out []Session
	err := s.client.do(ctx, http.MethodGet, sessionPath(sessionID, "children"), nil, nil, &out)
	return out, err
}

// Todos returns the task list of a session.
func (s *SessionService) Todos(ctx context.Context, sessionID string) ([]Todo, error) {
	var out []Todo
	err := s.client.do(ctx, http.MethodGet, sessionPath(sessionID, "todo"), nil, nil, &out)
	return out, err
}

// Status returns the status of all active sessions, keyed by session ID.
func (s *SessionService) Status(ctx context.Context) (map[string]SessionStatus, error) {
	var out map[string]SessionStatus
	err := s.client.do(ctx, http.MethodGet, "/session/status", nil, nil, &out)
	return out, err
}

// Diff returns the file differences accumulated in a session. messageID is
// optional; when set, the diff is computed up to that message.
func (s *SessionService) Diff(ctx context.Context, sessionID, messageID string) ([]SnapshotFileDiff, error) {
	q := url.Values{}
	if messageID != "" {
		q.Set("messageID", messageID)
	}
	var out []SnapshotFileDiff
	err := s.client.do(ctx, http.MethodGet, sessionPath(sessionID, "diff"), q, nil, &out)
	return out, err
}

// SessionInitParams configures session initialization.
type SessionInitParams struct {
	ProviderID string `json:"providerID"`
	ModelID    string `json:"modelID"`
	MessageID  string `json:"messageID"`
}

// Init analyzes the project and creates an AGENTS.md file.
func (s *SessionService) Init(ctx context.Context, sessionID string, params SessionInitParams) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "init"), nil, params, &out)
	return out, err
}

// Fork creates a new session branched from this one. messageID is optional;
// when set, the fork branches at that message.
func (s *SessionService) Fork(ctx context.Context, sessionID, messageID string) (*Session, error) {
	body := map[string]string{}
	if messageID != "" {
		body["messageID"] = messageID
	}
	var out Session
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "fork"), nil, body, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Abort stops the session's active run.
func (s *SessionService) Abort(ctx context.Context, sessionID string) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "abort"), nil, nil, &out)
	return out, err
}

// Share shares the session and returns it with the share URL set.
func (s *SessionService) Share(ctx context.Context, sessionID string) (*Session, error) {
	var out Session
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "share"), nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Unshare stops sharing the session.
func (s *SessionService) Unshare(ctx context.Context, sessionID string) (*Session, error) {
	var out Session
	err := s.client.do(ctx, http.MethodDelete, sessionPath(sessionID, "share"), nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SessionSummarizeParams configures conversation summarization.
type SessionSummarizeParams struct {
	ProviderID string `json:"providerID"`
	ModelID    string `json:"modelID"`
	Auto       bool   `json:"auto,omitempty"`
}

// Summarize summarizes (compacts) the session conversation.
func (s *SessionService) Summarize(ctx context.Context, sessionID string, params SessionSummarizeParams) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "summarize"), nil, params, &out)
	return out, err
}

// Revert reverts the session to the given message. partID is optional.
func (s *SessionService) Revert(ctx context.Context, sessionID, messageID, partID string) (*Session, error) {
	body := map[string]string{"messageID": messageID}
	if partID != "" {
		body["partID"] = partID
	}
	var out Session
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "revert"), nil, body, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Unrevert restores previously reverted messages.
func (s *SessionService) Unrevert(ctx context.Context, sessionID string) (*Session, error) {
	var out Session
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "unrevert"), nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PermissionResponse answers a permission request.
type PermissionResponse string

const (
	PermissionOnce   PermissionResponse = "once"
	PermissionAlways PermissionResponse = "always"
	PermissionReject PermissionResponse = "reject"
)

// RespondToPermission answers a pending permission request.
func (s *SessionService) RespondToPermission(ctx context.Context, sessionID, permissionID string, response PermissionResponse) (bool, error) {
	body := map[string]PermissionResponse{"response": response}
	var out bool
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "permissions", url.PathEscape(permissionID)), nil, body, &out)
	return out, err
}

// SessionMessagesParams paginates the message list.
type SessionMessagesParams struct {
	// Limit caps the number of returned messages.
	Limit int
	// Before returns messages before this message ID.
	Before string
}

// Messages lists the messages of a session. params may be nil.
func (s *SessionService) Messages(ctx context.Context, sessionID string, params *SessionMessagesParams) ([]MessageWithParts, error) {
	q := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Before != "" {
			q.Set("before", params.Before)
		}
	}
	var out []MessageWithParts
	err := s.client.do(ctx, http.MethodGet, sessionPath(sessionID, "message"), q, nil, &out)
	return out, err
}

// Message returns a single message with its parts.
func (s *SessionService) Message(ctx context.Context, sessionID, messageID string) (*MessageWithParts, error) {
	var out MessageWithParts
	err := s.client.do(ctx, http.MethodGet, sessionPath(sessionID, "message", url.PathEscape(messageID)), nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMessage deletes a message from a session.
func (s *SessionService) DeleteMessage(ctx context.Context, sessionID, messageID string) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodDelete, sessionPath(sessionID, "message", url.PathEscape(messageID)), nil, nil, &out)
	return out, err
}

// PromptParams is the body of a prompt (send message) request.
type PromptParams struct {
	MessageID string          `json:"messageID,omitempty"`
	Model     *ModelRef       `json:"model,omitempty"`
	Agent     string          `json:"agent,omitempty"`
	NoReply   bool            `json:"noReply,omitempty"`
	Tools     map[string]bool `json:"tools,omitempty"`
	Format    *OutputFormat   `json:"format,omitempty"`
	System    string          `json:"system,omitempty"`
	Variant   string          `json:"variant,omitempty"`
	Parts     []PartInput     `json:"parts"`
}

// Prompt sends a message to the session and waits for the assistant's reply.
func (s *SessionService) Prompt(ctx context.Context, sessionID string, params PromptParams) (*AssistantResponse, error) {
	var out AssistantResponse
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "message"), nil, params, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PromptText sends a plain text message and waits for the assistant's reply.
func (s *SessionService) PromptText(ctx context.Context, sessionID, text string) (*AssistantResponse, error) {
	return s.Prompt(ctx, sessionID, PromptParams{Parts: []PartInput{NewTextPart(text)}})
}

// PromptAsync sends a message without waiting for the reply (HTTP 204).
// Track progress via the event stream.
func (s *SessionService) PromptAsync(ctx context.Context, sessionID string, params PromptParams) error {
	return s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "prompt_async"), nil, params, nil)
}

// CommandParams executes a slash command in a session.
type CommandParams struct {
	MessageID string          `json:"messageID,omitempty"`
	Agent     string          `json:"agent,omitempty"`
	Model     string          `json:"model,omitempty"`
	Arguments string          `json:"arguments"`
	Command   string          `json:"command"`
	Variant   string          `json:"variant,omitempty"`
	Parts     []FilePartInput `json:"parts,omitempty"`
}

// Command executes a slash command and waits for the assistant's reply.
func (s *SessionService) Command(ctx context.Context, sessionID string, params CommandParams) (*AssistantResponse, error) {
	var out AssistantResponse
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "command"), nil, params, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ShellParams runs a shell command inside a session.
type ShellParams struct {
	MessageID string    `json:"messageID,omitempty"`
	Agent     string    `json:"agent"`
	Model     *ModelRef `json:"model,omitempty"`
	Command   string    `json:"command"`
}

// Shell runs a shell command inside the session and returns the resulting message.
func (s *SessionService) Shell(ctx context.Context, sessionID string, params ShellParams) (*MessageWithParts, error) {
	var out MessageWithParts
	err := s.client.do(ctx, http.MethodPost, sessionPath(sessionID, "shell"), nil, params, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

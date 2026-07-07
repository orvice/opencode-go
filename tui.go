package opencode

import (
	"context"
	"net/http"
)

// TUIService remote-controls an attached opencode TUI.
type TUIService struct {
	client *Client
}

// AppendPrompt appends text to the TUI prompt.
func (s *TUIService) AppendPrompt(ctx context.Context, text string) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, "/tui/append-prompt", nil, map[string]string{"text": text}, &out)
	return out, err
}

// SubmitPrompt submits the current TUI prompt.
func (s *TUIService) SubmitPrompt(ctx context.Context) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, "/tui/submit-prompt", nil, nil, &out)
	return out, err
}

// ClearPrompt clears the TUI prompt.
func (s *TUIService) ClearPrompt(ctx context.Context) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, "/tui/clear-prompt", nil, nil, &out)
	return out, err
}

// ExecuteCommand executes a TUI command (e.g. "agent_cycle").
func (s *TUIService) ExecuteCommand(ctx context.Context, command string) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, "/tui/execute-command", nil, map[string]string{"command": command}, &out)
	return out, err
}

// ToastParams configures a TUI toast notification.
type ToastParams struct {
	Title   string `json:"title,omitempty"`
	Message string `json:"message"`
	Variant string `json:"variant"` // info | success | warning | error
	// Duration in milliseconds.
	Duration int `json:"duration,omitempty"`
}

// ShowToast displays a toast notification in the TUI.
func (s *TUIService) ShowToast(ctx context.Context, params ToastParams) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, "/tui/show-toast", nil, params, &out)
	return out, err
}

// OpenHelp opens the help dialog.
func (s *TUIService) OpenHelp(ctx context.Context) (bool, error) {
	return s.open(ctx, "/tui/open-help")
}

// OpenSessions opens the session selector.
func (s *TUIService) OpenSessions(ctx context.Context) (bool, error) {
	return s.open(ctx, "/tui/open-sessions")
}

// OpenThemes opens the theme selector.
func (s *TUIService) OpenThemes(ctx context.Context) (bool, error) {
	return s.open(ctx, "/tui/open-themes")
}

// OpenModels opens the model selector.
func (s *TUIService) OpenModels(ctx context.Context) (bool, error) {
	return s.open(ctx, "/tui/open-models")
}

func (s *TUIService) open(ctx context.Context, path string) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, path, nil, nil, &out)
	return out, err
}

// TUIControlRequest is a pending control request from the TUI.
type TUIControlRequest struct {
	Path string `json:"path"`
	Body any    `json:"body"`
}

// ControlNext blocks until the TUI issues the next control request.
func (s *TUIService) ControlNext(ctx context.Context) (*TUIControlRequest, error) {
	var out TUIControlRequest
	err := s.client.do(ctx, http.MethodGet, "/tui/control/next", nil, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ControlResponse answers the control request obtained via ControlNext.
func (s *TUIService) ControlResponse(ctx context.Context, body any) (bool, error) {
	var out bool
	err := s.client.do(ctx, http.MethodPost, "/tui/control/response", nil, map[string]any{"body": body}, &out)
	return out, err
}

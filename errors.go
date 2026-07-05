package opencode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Error is returned when the server responds with a non-2xx status code.
type Error struct {
	StatusCode int
	Method     string
	Path       string
	// Message is the error message extracted from the response body, if any.
	Message string
	// Body is the raw response body.
	Body []byte
}

func (e *Error) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	return fmt.Sprintf("opencode: %s %s: %d %s", e.Method, e.Path, e.StatusCode, msg)
}

func newError(method, path string, resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	e := &Error{
		StatusCode: resp.StatusCode,
		Method:     method,
		Path:       path,
		Body:       body,
	}
	var payload struct {
		Message string `json:"message"`
		Data    struct {
			Message string `json:"message"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &payload) == nil {
		switch {
		case payload.Data.Message != "":
			e.Message = payload.Data.Message
		case payload.Message != "":
			e.Message = payload.Message
		}
	}
	if e.Message == "" && len(body) > 0 && len(body) <= 256 {
		e.Message = string(body)
	}
	return e
}

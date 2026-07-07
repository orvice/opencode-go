package opencode

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
)

// Event type constants for the most common bus events. The full set of
// event types is larger; unlisted events can still be received and decoded
// via Event.DecodeProperties.
const (
	EventTypeServerConnected    = "server.connected"
	EventTypeMessageUpdated     = "message.updated"
	EventTypeMessageRemoved     = "message.removed"
	EventTypeMessagePartUpdated = "message.part.updated"
	EventTypeMessagePartRemoved = "message.part.removed"
	EventTypeMessagePartDelta   = "message.part.delta"
	EventTypeSessionCreated     = "session.created"
	EventTypeSessionUpdated     = "session.updated"
	EventTypeSessionDeleted     = "session.deleted"
	EventTypeSessionIdle        = "session.idle"
	EventTypeSessionStatus      = "session.status"
	EventTypeSessionError       = "session.error"
	EventTypePermissionAsked    = "permission.asked"
	EventTypePermissionReplied  = "permission.replied"
	EventTypeFileEdited         = "file.edited"
	EventTypeTodoUpdated        = "todo.updated"
)

// Event is a server bus event. Properties holds the type-specific payload;
// decode it with DecodeProperties into one of the Event*Properties types
// (or any struct matching the payload).
type Event struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Properties json.RawMessage `json:"properties"`
}

// DecodeProperties unmarshals the event payload into v.
func (e *Event) DecodeProperties(v any) error {
	if len(e.Properties) == 0 {
		return nil
	}
	return json.Unmarshal(e.Properties, v)
}

// GlobalEvent is an event from the /global/event stream, which multiplexes
// events from all server instances.
type GlobalEvent struct {
	Directory string `json:"directory,omitempty"`
	Project   string `json:"project,omitempty"`
	Workspace string `json:"workspace,omitempty"`
	Payload   Event  `json:"payload"`
}

// EventMessageUpdatedProperties is the payload of "message.updated".
type EventMessageUpdatedProperties struct {
	SessionID string  `json:"sessionID"`
	Info      Message `json:"info"`
}

// EventMessageRemovedProperties is the payload of "message.removed".
type EventMessageRemovedProperties struct {
	SessionID string `json:"sessionID"`
	MessageID string `json:"messageID"`
}

// EventMessagePartUpdatedProperties is the payload of "message.part.updated".
type EventMessagePartUpdatedProperties struct {
	SessionID string `json:"sessionID"`
	Part      Part   `json:"part"`
	Time      int64  `json:"time"`
}

// EventMessagePartRemovedProperties is the payload of "message.part.removed".
type EventMessagePartRemovedProperties struct {
	SessionID string `json:"sessionID"`
	MessageID string `json:"messageID"`
	PartID    string `json:"partID"`
}

// EventMessagePartDeltaProperties is the payload of "message.part.delta",
// a streaming delta for the given field of a part (usually "text").
type EventMessagePartDeltaProperties struct {
	SessionID string `json:"sessionID"`
	MessageID string `json:"messageID"`
	PartID    string `json:"partID"`
	Field     string `json:"field"`
	Delta     string `json:"delta"`
}

// EventSessionProperties is the payload of "session.created",
// "session.updated" and "session.deleted".
type EventSessionProperties struct {
	SessionID string  `json:"sessionID"`
	Info      Session `json:"info"`
}

// EventSessionIdleProperties is the payload of "session.idle".
type EventSessionIdleProperties struct {
	SessionID string `json:"sessionID"`
}

// EventSessionStatusProperties is the payload of "session.status".
type EventSessionStatusProperties struct {
	SessionID string        `json:"sessionID"`
	Status    SessionStatus `json:"status"`
}

// EventSessionErrorProperties is the payload of "session.error".
type EventSessionErrorProperties struct {
	SessionID string        `json:"sessionID,omitempty"`
	Error     *MessageError `json:"error,omitempty"`
}

// EventFileEditedProperties is the payload of "file.edited".
type EventFileEditedProperties struct {
	File string `json:"file"`
}

// EventTodoUpdatedProperties is the payload of "todo.updated".
type EventTodoUpdatedProperties struct {
	SessionID string `json:"sessionID"`
	Todos     []Todo `json:"todos"`
}

// Stream is a server-sent event stream of JSON-encoded T values.
// Iterate with Next/Current and always Close when done:
//
//	stream, err := client.Event.Subscribe(ctx)
//	if err != nil { ... }
//	defer stream.Close()
//	for stream.Next() {
//		ev := stream.Current()
//		...
//	}
//	if err := stream.Err(); err != nil { ... }
type Stream[T any] struct {
	resp    *http.Response
	scanner *bufio.Scanner
	cur     T

	mu     sync.Mutex
	err    error
	closed bool
}

func (s *Stream[T]) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

func (s *Stream[T]) getErr() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

func (s *Stream[T]) setErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err == nil {
		s.err = err
	}
}

func newStream[T any](resp *http.Response) *Stream[T] {
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	return &Stream[T]{resp: resp, scanner: sc}
}

// Next advances to the next event. It blocks until an event arrives, the
// stream ends, or the request context is cancelled. Close may be called
// from another goroutine to unblock Next; Next and Current themselves must
// not be called concurrently.
func (s *Stream[T]) Next() bool {
	if s.getErr() != nil || s.isClosed() {
		return false
	}
	var data []byte
	emit := func() bool {
		var v T
		if err := json.Unmarshal(data, &v); err != nil {
			s.setErr(err)
			return false
		}
		s.cur = v
		return true
	}
	for s.scanner.Scan() {
		line := s.scanner.Bytes()
		switch {
		case len(line) == 0:
			if len(data) > 0 {
				return emit()
			}
		case bytes.HasPrefix(line, []byte("data:")):
			payload := bytes.TrimPrefix(line, []byte("data:"))
			payload = bytes.TrimPrefix(payload, []byte(" "))
			if len(data) > 0 {
				data = append(data, '\n')
			}
			data = append(data, payload...)
		default:
			// Ignore other SSE fields (event:, id:, retry:) and comments.
		}
	}
	if err := s.scanner.Err(); err != nil && !s.isClosed() {
		s.setErr(err)
		return false
	}
	if len(data) > 0 {
		return emit()
	}
	return false
}

// Current returns the event read by the last successful Next call.
func (s *Stream[T]) Current() T { return s.cur }

// Err returns the error that ended the stream, if any. A nil return after
// Next returns false means the stream ended normally (or was closed).
func (s *Stream[T]) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	return s.err
}

// Close terminates the stream and releases the underlying connection.
// It is safe to call Close from another goroutine while Next is blocked.
func (s *Stream[T]) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()
	return s.resp.Body.Close()
}

func subscribe[T any](ctx context.Context, c *Client, path string) (*Stream[T], error) {
	req, err := c.newRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, newError(http.MethodGet, path, resp)
	}
	return newStream[T](resp), nil
}

// EventService subscribes to server-sent events.
type EventService struct {
	client *Client
}

// Subscribe opens the event stream at /event for the current instance.
// The first event received is always "server.connected".
func (s *EventService) Subscribe(ctx context.Context) (*Stream[Event], error) {
	return subscribe[Event](ctx, s.client, "/event")
}

// Events opens the global event stream at /global/event, which carries
// events from all server instances tagged with directory/project/workspace.
func (s *GlobalService) Events(ctx context.Context) (*Stream[GlobalEvent], error) {
	return subscribe[GlobalEvent](ctx, s.client, "/global/event")
}

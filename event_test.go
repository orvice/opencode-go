package opencode

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestEventStream(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/event" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		f := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"id\":\"1\",\"type\":\"server.connected\",\"properties\":{}}\n\n")
		f.Flush()
		fmt.Fprint(w, ": keepalive comment\n")
		fmt.Fprint(w, "data: {\"id\":\"2\",\"type\":\"session.idle\",\"properties\":{\"sessionID\":\"ses_1\"}}\n\n")
		f.Flush()
	}))

	stream, err := c.Event.Subscribe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	var events []Event
	for stream.Next() {
		events = append(events, stream.Current())
	}
	if err := stream.Err(); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("events = %+v", events)
	}
	if events[0].Type != EventTypeServerConnected {
		t.Errorf("first event = %+v", events[0])
	}
	var props EventSessionIdleProperties
	if err := events[1].DecodeProperties(&props); err != nil {
		t.Fatal(err)
	}
	if props.SessionID != "ses_1" {
		t.Errorf("props = %+v", props)
	}
}

func TestEventStreamContextCancel(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		f := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"id\":\"1\",\"type\":\"server.connected\",\"properties\":{}}\n\n")
		f.Flush()
		<-r.Context().Done()
	}))

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := c.Event.Subscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	if !stream.Next() {
		t.Fatalf("expected first event, err = %v", stream.Err())
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	done := make(chan struct{})
	go func() {
		for stream.Next() {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("stream did not terminate after context cancel")
	}
}

func TestEventStreamCloseDuringNext(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		f := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"id\":\"1\",\"type\":\"server.connected\",\"properties\":{}}\n\n")
		f.Flush()
		<-r.Context().Done()
	}))

	stream, err := c.Event.Subscribe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !stream.Next() {
		t.Fatalf("expected first event, err = %v", stream.Err())
	}

	// Close from another goroutine while Next is blocked; this must unblock
	// Next and must not race (run with -race).
	go func() {
		time.Sleep(50 * time.Millisecond)
		stream.Close()
	}()
	done := make(chan struct{})
	go func() {
		for stream.Next() {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("stream did not terminate after Close")
	}
	if err := stream.Err(); err != nil {
		t.Errorf("Err() after Close = %v, want nil", err)
	}
}

func TestGlobalEventStream(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/global/event" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"directory\":\"/tmp\",\"payload\":{\"id\":\"1\",\"type\":\"server.connected\",\"properties\":{}}}\n\n")
	}))

	stream, err := c.Global.Events(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	if !stream.Next() {
		t.Fatalf("no event, err = %v", stream.Err())
	}
	ev := stream.Current()
	if ev.Directory != "/tmp" || ev.Payload.Type != EventTypeServerConnected {
		t.Errorf("event = %+v", ev)
	}
}

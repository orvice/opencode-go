package opencode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.Handler, opts ...Option) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient(append([]Option{WithBaseURL(srv.URL)}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestHealth(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/global/health" {
			t.Errorf("path = %q, want /global/health", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"healthy": true, "version": "1.15.13"})
	}))
	h, err := c.Global.Health(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !h.Healthy || h.Version != "1.15.13" {
		t.Errorf("health = %+v", h)
	}
}

func TestBasicAuthAndDefaultQuery(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "opencode" || pass != "secret" {
			t.Errorf("basic auth = %q %q %v", user, pass, ok)
		}
		if got := r.URL.Query().Get("directory"); got != "/tmp/project" {
			t.Errorf("directory = %q", got)
		}
		if got := r.URL.Query().Get("workspace"); got != "wrk_1" {
			t.Errorf("workspace = %q", got)
		}
		w.Write([]byte("[]"))
	}), WithPassword("secret"), WithDirectory("/tmp/project"), WithWorkspace("wrk_1"))

	if _, err := c.Session.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestErrorParsing(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"data":{"message":"session not found"}}`))
	}))
	_, err := c.Session.Get(context.Background(), "ses_missing")
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("err = %T (%v), want *Error", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d", apiErr.StatusCode)
	}
	if apiErr.Message != "session not found" {
		t.Errorf("message = %q", apiErr.Message)
	}
}

func TestNoContentResponse(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	err := c.Session.PromptAsync(context.Background(), "ses_1", PromptParams{
		Parts: []PartInput{NewTextPart("hi")},
	})
	if err != nil {
		t.Fatal(err)
	}
}

package opencode

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestIntegration exercises the SDK against a real opencode server.
// It is skipped unless OPENCODE_INTEGRATION_URL is set, e.g.:
//
//	opencode serve --port 4096 &
//	OPENCODE_INTEGRATION_URL=http://127.0.0.1:4096 go test -run TestIntegration -v
func TestIntegration(t *testing.T) {
	baseURL := os.Getenv("OPENCODE_INTEGRATION_URL")
	if baseURL == "" {
		t.Skip("OPENCODE_INTEGRATION_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	c, err := NewClient(WithBaseURL(baseURL))
	if err != nil {
		t.Fatal(err)
	}

	health, err := c.Global.Health(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !health.Healthy {
		t.Fatalf("health = %+v", health)
	}
	t.Logf("server version %s", health.Version)

	// Event stream: first event must be server.connected.
	stream, err := c.Event.Subscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if !stream.Next() {
		t.Fatalf("no event received: %v", stream.Err())
	}
	if got := stream.Current().Type; got != EventTypeServerConnected {
		t.Errorf("first event = %q, want %q", got, EventTypeServerConnected)
	}

	// Session lifecycle.
	sess, err := c.Session.Create(ctx, &SessionCreateParams{Title: "opencode-go integration test"})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("created session %s", sess.ID)

	got, err := c.Session.Get(ctx, sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != sess.ID {
		t.Errorf("get = %+v", got)
	}

	updated, err := c.Session.Update(ctx, sess.ID, &SessionUpdateParams{Title: "renamed"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "renamed" {
		t.Errorf("update title = %q", updated.Title)
	}

	if _, err := c.Session.Messages(ctx, sess.ID, nil); err != nil {
		t.Errorf("messages: %v", err)
	}

	if ok, err := c.Session.Delete(ctx, sess.ID); err != nil || !ok {
		t.Errorf("delete: ok=%v err=%v", ok, err)
	}

	// Read-only surfaces.
	if _, err := c.App.Agents(ctx); err != nil {
		t.Errorf("agents: %v", err)
	}
	if _, err := c.App.Commands(ctx); err != nil {
		t.Errorf("commands: %v", err)
	}
	if _, err := c.App.Path(ctx); err != nil {
		t.Errorf("path: %v", err)
	}
	if _, err := c.Project.Current(ctx); err != nil {
		t.Errorf("project current: %v", err)
	}
	if _, err := c.Config.Get(ctx); err != nil {
		t.Errorf("config: %v", err)
	}
	if _, err := c.Provider.List(ctx); err != nil {
		t.Errorf("providers: %v", err)
	}
	if _, err := c.File.Status(ctx); err != nil {
		t.Errorf("file status: %v", err)
	}
	if _, err := c.Session.Status(ctx); err != nil {
		t.Errorf("session status: %v", err)
	}
}

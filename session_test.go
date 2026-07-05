package opencode

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSessionCreateAndPrompt(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /session", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["title"] != "my session" {
			t.Errorf("title = %v", body["title"])
		}
		json.NewEncoder(w).Encode(Session{
			ID: "ses_1", Slug: "my-session", ProjectID: "prj_1",
			Directory: "/tmp", Title: "my session", Version: "1.15.13",
			Time: SessionTime{Created: 1, Updated: 1},
		})
	})
	mux.HandleFunc("POST /session/ses_1/message", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Parts []map[string]any `json:"parts"`
			Model *ModelRef        `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body.Parts) != 1 || body.Parts[0]["text"] != "hello" {
			t.Errorf("parts = %v", body.Parts)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"info": AssistantMessage{
				ID: "msg_2", SessionID: "ses_1", Role: RoleAssistant,
				ParentID: "msg_1", ModelID: "m", ProviderID: "p", Mode: "build", Agent: "build",
			},
			"parts": []Part{{ID: "prt_1", SessionID: "ses_1", MessageID: "msg_2", Type: PartTypeText, Text: "hi there"}},
		})
	})

	c := newTestClient(t, mux)
	ctx := context.Background()

	sess, err := c.Session.Create(ctx, &SessionCreateParams{Title: "my session"})
	if err != nil {
		t.Fatal(err)
	}
	if sess.ID != "ses_1" {
		t.Fatalf("session = %+v", sess)
	}

	resp, err := c.Session.PromptText(ctx, sess.ID, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Info.ID != "msg_2" || resp.Text() != "hi there" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestSessionListParams(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("scope") != "project" || q.Get("limit") != "5" || q.Get("roots") != "true" || q.Get("search") != "abc" {
			t.Errorf("query = %v", q)
		}
		w.Write([]byte("[]"))
	}))
	roots := true
	_, err := c.Session.List(context.Background(), &SessionListParams{
		Scope: "project", Limit: 5, Roots: &roots, Search: "abc",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSessionMessages(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/session/ses_1/message" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("limit = %q", got)
		}
		w.Write([]byte(`[
			{"info":{"id":"msg_1","sessionID":"ses_1","role":"user","time":{"created":1},
			         "agent":"build","model":{"providerID":"p","modelID":"m"}},
			 "parts":[{"id":"prt_1","sessionID":"ses_1","messageID":"msg_1","type":"text","text":"q"}]}
		]`))
	}))
	msgs, err := c.Session.Messages(context.Background(), "ses_1", &SessionMessagesParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0].Info.Role() != RoleUser || msgs[0].Parts[0].Text != "q" {
		t.Errorf("msgs = %+v", msgs)
	}
}

func TestRespondToPermission(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/session/ses_1/permissions/per_1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["response"] != "always" {
			t.Errorf("body = %v", body)
		}
		w.Write([]byte("true"))
	}))
	ok, err := c.Session.RespondToPermission(context.Background(), "ses_1", "per_1", PermissionAlways)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

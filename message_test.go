package opencode

import (
	"encoding/json"
	"testing"
)

func TestMessageUnmarshalUser(t *testing.T) {
	raw := `{
		"id": "msg_1", "sessionID": "ses_1", "role": "user",
		"time": {"created": 1751700000000},
		"agent": "build",
		"model": {"providerID": "anthropic", "modelID": "claude-sonnet-4-6"}
	}`
	var m Message
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	if m.Role() != RoleUser || m.User == nil || m.Assistant != nil {
		t.Fatalf("m = %+v", m)
	}
	if m.User.Model.ProviderID != "anthropic" || m.ID() != "msg_1" {
		t.Errorf("user = %+v", m.User)
	}
}

func TestMessageUnmarshalAssistant(t *testing.T) {
	raw := `{
		"id": "msg_2", "sessionID": "ses_1", "role": "assistant",
		"time": {"created": 1751700000000, "completed": 1751700001000},
		"parentID": "msg_1", "modelID": "claude-sonnet-4-6", "providerID": "anthropic",
		"mode": "build", "agent": "build",
		"path": {"cwd": "/tmp", "root": "/tmp"},
		"cost": 0.01,
		"tokens": {"input": 10, "output": 20, "reasoning": 0, "cache": {"read": 0, "write": 0}},
		"error": {"name": "APIError", "data": {"message": "rate limited", "statusCode": 429, "isRetryable": true}}
	}`
	var m Message
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	a := m.Assistant
	if m.Role() != RoleAssistant || a == nil {
		t.Fatalf("m = %+v", m)
	}
	if a.Tokens.Output != 20 || a.Error.Name != "APIError" || a.Error.Data.StatusCode != 429 {
		t.Errorf("assistant = %+v", a)
	}
}

func TestMessageMarshalRoundTrip(t *testing.T) {
	m := Message{User: &UserMessage{ID: "msg_1", SessionID: "ses_1", Role: RoleUser, Agent: "build"}}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var back Message
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.User == nil || back.User.ID != "msg_1" {
		t.Errorf("round trip = %+v", back)
	}
}

func TestPartUnionDecoding(t *testing.T) {
	raw := `[
		{"id":"prt_1","sessionID":"ses_1","messageID":"msg_2","type":"text","text":"hello"},
		{"id":"prt_2","sessionID":"ses_1","messageID":"msg_2","type":"tool","callID":"c1","tool":"bash",
		 "state":{"status":"completed","input":{"command":"ls"},"output":"ok","title":"ls","metadata":{},
		          "time":{"start":1,"end":2}}},
		{"id":"prt_3","sessionID":"ses_1","messageID":"msg_2","type":"step-finish","reason":"stop","cost":0.1,
		 "tokens":{"input":1,"output":2,"reasoning":0,"cache":{"read":0,"write":0}}},
		{"id":"prt_4","sessionID":"ses_1","messageID":"msg_2","type":"file","mime":"image/png",
		 "url":"data:image/png;base64,xxx","source":{"type":"file","path":"a.png","text":{"value":"","start":0,"end":0}}}
	]`
	var parts []Part
	if err := json.Unmarshal([]byte(raw), &parts); err != nil {
		t.Fatal(err)
	}
	if parts[0].Type != PartTypeText || parts[0].Text != "hello" {
		t.Errorf("text part = %+v", parts[0])
	}
	if parts[1].State == nil || parts[1].State.Status != ToolStatusCompleted || parts[1].State.Output != "ok" {
		t.Errorf("tool part = %+v", parts[1])
	}
	if parts[2].Tokens == nil || parts[2].Tokens.Output != 2 {
		t.Errorf("step-finish part = %+v", parts[2])
	}
	if parts[3].Source == nil || parts[3].Source.Path != "a.png" {
		t.Errorf("file part = %+v", parts[3])
	}
}

func TestPartInputMarshal(t *testing.T) {
	params := PromptParams{
		Model: &ModelRef{ProviderID: "anthropic", ModelID: "claude-sonnet-4-6"},
		Parts: []PartInput{
			NewTextPart("hi"),
			NewFilePart("image/png", "data:image/png;base64,xxx"),
			NewAgentPart("plan"),
		},
	}
	b, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Model ModelRef `json:"model"`
		Parts []map[string]any
	}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Parts) != 3 {
		t.Fatalf("parts = %v", decoded.Parts)
	}
	if decoded.Parts[0]["type"] != "text" || decoded.Parts[0]["text"] != "hi" {
		t.Errorf("text input = %v", decoded.Parts[0])
	}
	if decoded.Parts[1]["mime"] != "image/png" {
		t.Errorf("file input = %v", decoded.Parts[1])
	}
	if decoded.Parts[2]["name"] != "plan" {
		t.Errorf("agent input = %v", decoded.Parts[2])
	}
}

func TestAssistantResponseText(t *testing.T) {
	r := AssistantResponse{Parts: []Part{
		{Type: PartTypeStepStart},
		{Type: PartTypeText, Text: "hello "},
		{Type: PartTypeText, Text: "world"},
	}}
	if got := r.Text(); got != "hello world" {
		t.Errorf("Text() = %q", got)
	}
}

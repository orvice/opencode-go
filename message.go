package opencode

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MessageRole is the role of a message author.
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// ModelRef identifies a model by provider and model ID.
type ModelRef struct {
	ProviderID string `json:"providerID"`
	ModelID    string `json:"modelID"`
}

// OutputFormat controls the response format of a prompt.
// Type is "text" (default) or "json_schema"; Schema and RetryCount apply
// to "json_schema" only.
type OutputFormat struct {
	Type       string         `json:"type"`
	Schema     map[string]any `json:"schema,omitempty"`
	RetryCount int            `json:"retryCount,omitempty"`
}

// MessageErrorData is the payload of a MessageError. Fields are populated
// depending on the error name.
type MessageErrorData struct {
	Message         string            `json:"message,omitempty"`
	ProviderID      string            `json:"providerID,omitempty"`
	StatusCode      int               `json:"statusCode,omitempty"`
	IsRetryable     bool              `json:"isRetryable,omitempty"`
	ResponseHeaders map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody    string            `json:"responseBody,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Retries         int               `json:"retries,omitempty"`
	Ref             string            `json:"ref,omitempty"`
}

// MessageError is an error attached to an assistant message or session.
// Name is one of "ProviderAuthError", "UnknownError", "MessageOutputLengthError",
// "MessageAbortedError", "StructuredOutputError", "ContextOverflowError" or "APIError".
type MessageError struct {
	Name string           `json:"name"`
	Data MessageErrorData `json:"data"`
}

// UserMessageTime holds user message timestamps (unix milliseconds).
type UserMessageTime struct {
	Created int64 `json:"created"`
}

// UserMessageSummary is a summary generated for a user message.
type UserMessageSummary struct {
	Title string             `json:"title,omitempty"`
	Body  string             `json:"body,omitempty"`
	Diffs []SnapshotFileDiff `json:"diffs"`
}

// UserMessageModel identifies the model targeted by a user message.
type UserMessageModel struct {
	ProviderID string `json:"providerID"`
	ModelID    string `json:"modelID"`
	Variant    string `json:"variant,omitempty"`
}

// UserMessage is a message authored by the user.
type UserMessage struct {
	ID        string              `json:"id"`
	SessionID string              `json:"sessionID"`
	Role      MessageRole         `json:"role"`
	Time      UserMessageTime     `json:"time"`
	Format    *OutputFormat       `json:"format,omitempty"`
	Summary   *UserMessageSummary `json:"summary,omitempty"`
	Agent     string              `json:"agent"`
	Model     UserMessageModel    `json:"model"`
	System    string              `json:"system,omitempty"`
	Tools     map[string]bool     `json:"tools,omitempty"`
}

// AssistantMessageTime holds assistant message timestamps (unix milliseconds).
type AssistantMessageTime struct {
	Created   int64 `json:"created"`
	Completed int64 `json:"completed,omitempty"`
}

// AssistantMessagePath records the working directory of an assistant message.
type AssistantMessagePath struct {
	Cwd  string `json:"cwd"`
	Root string `json:"root"`
}

// AssistantMessage is a message authored by the assistant.
type AssistantMessage struct {
	ID         string               `json:"id"`
	SessionID  string               `json:"sessionID"`
	Role       MessageRole          `json:"role"`
	Time       AssistantMessageTime `json:"time"`
	Error      *MessageError        `json:"error,omitempty"`
	ParentID   string               `json:"parentID"`
	ModelID    string               `json:"modelID"`
	ProviderID string               `json:"providerID"`
	Mode       string               `json:"mode"`
	Agent      string               `json:"agent"`
	Path       AssistantMessagePath `json:"path"`
	Summary    bool                 `json:"summary,omitempty"`
	Cost       float64              `json:"cost"`
	Tokens     TokenUsage           `json:"tokens"`
	Structured any                  `json:"structured,omitempty"`
	Variant    string               `json:"variant,omitempty"`
	Finish     string               `json:"finish,omitempty"`
}

// Message is a union of UserMessage and AssistantMessage.
// Exactly one of User and Assistant is non-nil.
type Message struct {
	User      *UserMessage
	Assistant *AssistantMessage
}

// Role returns the role of the message.
func (m Message) Role() MessageRole {
	switch {
	case m.User != nil:
		return RoleUser
	case m.Assistant != nil:
		return RoleAssistant
	}
	return ""
}

// ID returns the message ID regardless of role.
func (m Message) ID() string {
	switch {
	case m.User != nil:
		return m.User.ID
	case m.Assistant != nil:
		return m.Assistant.ID
	}
	return ""
}

func (m *Message) UnmarshalJSON(b []byte) error {
	var probe struct {
		Role MessageRole `json:"role"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		return err
	}
	switch probe.Role {
	case RoleUser:
		m.Assistant = nil
		m.User = new(UserMessage)
		return json.Unmarshal(b, m.User)
	case RoleAssistant:
		m.User = nil
		m.Assistant = new(AssistantMessage)
		return json.Unmarshal(b, m.Assistant)
	default:
		return fmt.Errorf("opencode: unknown message role %q", probe.Role)
	}
}

func (m Message) MarshalJSON() ([]byte, error) {
	switch {
	case m.User != nil:
		return json.Marshal(m.User)
	case m.Assistant != nil:
		return json.Marshal(m.Assistant)
	}
	return []byte("null"), nil
}

// PartType discriminates message part variants.
type PartType string

const (
	PartTypeText       PartType = "text"
	PartTypeSubtask    PartType = "subtask"
	PartTypeReasoning  PartType = "reasoning"
	PartTypeFile       PartType = "file"
	PartTypeTool       PartType = "tool"
	PartTypeStepStart  PartType = "step-start"
	PartTypeStepFinish PartType = "step-finish"
	PartTypeSnapshot   PartType = "snapshot"
	PartTypePatch      PartType = "patch"
	PartTypeAgent      PartType = "agent"
	PartTypeRetry      PartType = "retry"
	PartTypeCompaction PartType = "compaction"
)

// PartTime holds part timestamps (unix milliseconds). Which fields are set
// depends on the part type.
type PartTime struct {
	Start   int64 `json:"start,omitempty"`
	End     int64 `json:"end,omitempty"`
	Created int64 `json:"created,omitempty"`
}

// PartSource describes where a file/agent part came from. It merges the
// "file", "symbol", "resource" and agent-mention source variants; Type is
// empty for agent mention sources.
type PartSource struct {
	Type string          `json:"type,omitempty"` // file | symbol | resource
	Path string          `json:"path,omitempty"`
	Text *PartSourceText `json:"text,omitempty"`
	// Symbol source fields.
	Range *Range `json:"range,omitempty"`
	Name  string `json:"name,omitempty"`
	Kind  int    `json:"kind,omitempty"`
	// Resource source fields.
	ClientName string `json:"clientName,omitempty"`
	URI        string `json:"uri,omitempty"`
	// Agent mention source fields.
	Value string `json:"value,omitempty"`
	Start int    `json:"start,omitempty"`
	End   int    `json:"end,omitempty"`
}

// PartSourceText is the text range a part source refers to.
type PartSourceText struct {
	Value string  `json:"value"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// ToolStatus is the state of a tool call.
type ToolStatus string

const (
	ToolStatusPending   ToolStatus = "pending"
	ToolStatusRunning   ToolStatus = "running"
	ToolStatusCompleted ToolStatus = "completed"
	ToolStatusError     ToolStatus = "error"
)

// ToolStateTime holds tool execution timestamps (unix milliseconds).
type ToolStateTime struct {
	Start     int64 `json:"start,omitempty"`
	End       int64 `json:"end,omitempty"`
	Compacted int64 `json:"compacted,omitempty"`
}

// ToolState is the state of a tool call. Fields are populated depending on
// Status: Raw for pending; Title from running on; Output and Attachments on
// completed; Error on error.
type ToolState struct {
	Status      ToolStatus     `json:"status"`
	Input       map[string]any `json:"input"`
	Raw         string         `json:"raw,omitempty"`
	Title       string         `json:"title,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Output      string         `json:"output,omitempty"`
	Error       string         `json:"error,omitempty"`
	Time        *ToolStateTime `json:"time,omitempty"`
	Attachments []Part         `json:"attachments,omitempty"`
}

// Part is a message part. It is a flattened union discriminated by Type;
// only the fields relevant to the given type are populated.
type Part struct {
	ID        string   `json:"id"`
	SessionID string   `json:"sessionID"`
	MessageID string   `json:"messageID"`
	Type      PartType `json:"type"`

	// Text and reasoning parts.
	Text      string         `json:"text,omitempty"`
	Synthetic bool           `json:"synthetic,omitempty"`
	Ignored   bool           `json:"ignored,omitempty"`
	Time      *PartTime      `json:"time,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`

	// File parts.
	Mime     string      `json:"mime,omitempty"`
	Filename string      `json:"filename,omitempty"`
	URL      string      `json:"url,omitempty"`
	Source   *PartSource `json:"source,omitempty"`

	// Tool parts.
	CallID string     `json:"callID,omitempty"`
	Tool   string     `json:"tool,omitempty"`
	State  *ToolState `json:"state,omitempty"`

	// Step and snapshot parts.
	Snapshot string      `json:"snapshot,omitempty"`
	Reason   string      `json:"reason,omitempty"`
	Cost     float64     `json:"cost,omitempty"`
	Tokens   *TokenUsage `json:"tokens,omitempty"`

	// Patch parts.
	Hash  string   `json:"hash,omitempty"`
	Files []string `json:"files,omitempty"`

	// Agent parts.
	Name string `json:"name,omitempty"`

	// Retry parts.
	Attempt int           `json:"attempt,omitempty"`
	Error   *MessageError `json:"error,omitempty"`

	// Compaction parts.
	Auto        bool   `json:"auto,omitempty"`
	Overflow    bool   `json:"overflow,omitempty"`
	TailStartID string `json:"tail_start_id,omitempty"`

	// Subtask parts.
	Prompt      string    `json:"prompt,omitempty"`
	Description string    `json:"description,omitempty"`
	Agent       string    `json:"agent,omitempty"`
	Model       *ModelRef `json:"model,omitempty"`
	Command     string    `json:"command,omitempty"`
}

// MessageWithParts pairs a message with its parts, as returned by the
// message listing and detail endpoints.
type MessageWithParts struct {
	Info  Message `json:"info"`
	Parts []Part  `json:"parts"`
}

// AssistantResponse is the assistant reply returned by prompt, command and
// shell endpoints.
type AssistantResponse struct {
	Info  AssistantMessage `json:"info"`
	Parts []Part           `json:"parts"`
}

// Text concatenates all text parts of the response.
func (r *AssistantResponse) Text() string {
	var sb strings.Builder
	for _, p := range r.Parts {
		if p.Type == PartTypeText {
			sb.WriteString(p.Text)
		}
	}
	return sb.String()
}

// PartInput is implemented by the part types that can be sent in a prompt:
// TextPartInput, FilePartInput, AgentPartInput and SubtaskPartInput.
type PartInput interface {
	isPartInput()
}

// TextPartInput is a text part of an outgoing prompt.
type TextPartInput struct {
	ID        string         `json:"id,omitempty"`
	Type      PartType       `json:"type"` // must be PartTypeText
	Text      string         `json:"text"`
	Synthetic bool           `json:"synthetic,omitempty"`
	Ignored   bool           `json:"ignored,omitempty"`
	Time      *PartTime      `json:"time,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func (TextPartInput) isPartInput() {}

// NewTextPart returns a text part for an outgoing prompt.
func NewTextPart(text string) TextPartInput {
	return TextPartInput{Type: PartTypeText, Text: text}
}

// FilePartInput is a file attachment part of an outgoing prompt.
type FilePartInput struct {
	ID       string      `json:"id,omitempty"`
	Type     PartType    `json:"type"` // must be PartTypeFile
	Mime     string      `json:"mime"`
	Filename string      `json:"filename,omitempty"`
	URL      string      `json:"url"`
	Source   *PartSource `json:"source,omitempty"`
}

func (FilePartInput) isPartInput() {}

// NewFilePart returns a file part for an outgoing prompt. The URL may be a
// data URL or a file:// URL understood by the server.
func NewFilePart(mime, url string) FilePartInput {
	return FilePartInput{Type: PartTypeFile, Mime: mime, URL: url}
}

// AgentMentionSource records the text range of an @agent mention.
type AgentMentionSource struct {
	Value string `json:"value"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// AgentPartInput mentions an agent in an outgoing prompt.
type AgentPartInput struct {
	ID     string              `json:"id,omitempty"`
	Type   PartType            `json:"type"` // must be PartTypeAgent
	Name   string              `json:"name"`
	Source *AgentMentionSource `json:"source,omitempty"`
}

func (AgentPartInput) isPartInput() {}

// NewAgentPart returns an agent mention part for an outgoing prompt.
func NewAgentPart(name string) AgentPartInput {
	return AgentPartInput{Type: PartTypeAgent, Name: name}
}

// SubtaskPartInput requests a subtask in an outgoing prompt.
type SubtaskPartInput struct {
	ID          string    `json:"id,omitempty"`
	Type        PartType  `json:"type"` // must be PartTypeSubtask
	Prompt      string    `json:"prompt"`
	Description string    `json:"description"`
	Agent       string    `json:"agent"`
	Model       *ModelRef `json:"model,omitempty"`
	Command     string    `json:"command,omitempty"`
}

func (SubtaskPartInput) isPartInput() {}

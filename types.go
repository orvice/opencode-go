package opencode

// TokenCache reports cached token counts.
type TokenCache struct {
	Read  float64 `json:"read"`
	Write float64 `json:"write"`
}

// TokenUsage reports token consumption for a session, message or step.
type TokenUsage struct {
	Total     float64    `json:"total,omitempty"`
	Input     float64    `json:"input"`
	Output    float64    `json:"output"`
	Reasoning float64    `json:"reasoning"`
	Cache     TokenCache `json:"cache"`
}

// SnapshotFileDiff describes a per-file diff in a session snapshot.
type SnapshotFileDiff struct {
	File      string  `json:"file,omitempty"`
	Patch     string  `json:"patch,omitempty"`
	Additions float64 `json:"additions"`
	Deletions float64 `json:"deletions"`
	Status    string  `json:"status,omitempty"` // "added", "deleted" or "modified"
}

// SessionTime holds session timestamps (unix milliseconds).
type SessionTime struct {
	Created    int64 `json:"created"`
	Updated    int64 `json:"updated"`
	Compacting int64 `json:"compacting,omitempty"`
	Archived   int64 `json:"archived,omitempty"`
}

// SessionSummary aggregates file changes made in a session.
type SessionSummary struct {
	Additions float64            `json:"additions"`
	Deletions float64            `json:"deletions"`
	Files     float64            `json:"files"`
	Diffs     []SnapshotFileDiff `json:"diffs,omitempty"`
}

// SessionShare holds the public share URL of a shared session.
type SessionShare struct {
	URL string `json:"url"`
}

// SessionModel identifies the model a session uses.
type SessionModel struct {
	ID         string `json:"id"`
	ProviderID string `json:"providerID"`
	Variant    string `json:"variant,omitempty"`
}

// SessionRevert records the revert point of a session.
type SessionRevert struct {
	MessageID string `json:"messageID"`
	PartID    string `json:"partID,omitempty"`
	Snapshot  string `json:"snapshot,omitempty"`
	Diff      string `json:"diff,omitempty"`
}

// Session is an opencode conversation session.
type Session struct {
	ID          string            `json:"id"`
	Slug        string            `json:"slug"`
	ProjectID   string            `json:"projectID"`
	WorkspaceID string            `json:"workspaceID,omitempty"`
	Directory   string            `json:"directory"`
	Path        string            `json:"path,omitempty"`
	ParentID    string            `json:"parentID,omitempty"`
	Summary     *SessionSummary   `json:"summary,omitempty"`
	Cost        float64           `json:"cost,omitempty"`
	Tokens      *TokenUsage       `json:"tokens,omitempty"`
	Share       *SessionShare     `json:"share,omitempty"`
	Title       string            `json:"title"`
	Agent       string            `json:"agent,omitempty"`
	Model       *SessionModel     `json:"model,omitempty"`
	Version     string            `json:"version"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
	Time        SessionTime       `json:"time"`
	Permission  PermissionRuleset `json:"permission,omitempty"`
	Revert      *SessionRevert    `json:"revert,omitempty"`
}

// SessionStatusType discriminates SessionStatus values.
type SessionStatusType string

const (
	SessionStatusIdle  SessionStatusType = "idle"
	SessionStatusRetry SessionStatusType = "retry"
	SessionStatusBusy  SessionStatusType = "busy"
)

// SessionStatus is the current status of a session.
// Attempt, Message, Action and Next are only set when Type is "retry".
type SessionStatus struct {
	Type    SessionStatusType `json:"type"`
	Attempt int               `json:"attempt,omitempty"`
	Message string            `json:"message,omitempty"`
	Action  *RetryAction      `json:"action,omitempty"`
	Next    int64             `json:"next,omitempty"`
}

// RetryAction describes a user action suggested for a retry status.
type RetryAction struct {
	Reason   string `json:"reason"`
	Provider string `json:"provider"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Label    string `json:"label"`
	Link     string `json:"link,omitempty"`
}

// PermissionAction is the effect of a permission rule.
type PermissionAction string

const (
	PermissionAllow PermissionAction = "allow"
	PermissionDeny  PermissionAction = "deny"
	PermissionAsk   PermissionAction = "ask"
)

// PermissionRule is a single permission rule.
type PermissionRule struct {
	Permission string           `json:"permission"`
	Pattern    string           `json:"pattern"`
	Action     PermissionAction `json:"action"`
}

// PermissionRuleset is an ordered list of permission rules.
type PermissionRuleset []PermissionRule

// PermissionRequestTool references the tool call that triggered a permission request.
type PermissionRequestTool struct {
	MessageID string `json:"messageID"`
	CallID    string `json:"callID"`
}

// PermissionRequest is a pending permission request from the server.
type PermissionRequest struct {
	ID         string                 `json:"id"`
	SessionID  string                 `json:"sessionID"`
	Permission string                 `json:"permission"`
	Patterns   []string               `json:"patterns"`
	Metadata   map[string]any         `json:"metadata"`
	Always     []string               `json:"always"`
	Tool       *PermissionRequestTool `json:"tool,omitempty"`
}

// ModelAPI describes the upstream API used by a model.
type ModelAPI struct {
	ID  string `json:"id"`
	URL string `json:"url"`
	NPM string `json:"npm"`
}

// ModelIO lists supported input/output modalities.
type ModelIO struct {
	Text  bool `json:"text"`
	Audio bool `json:"audio"`
	Image bool `json:"image"`
	Video bool `json:"video"`
	PDF   bool `json:"pdf"`
}

// ModelCapabilities describes what a model supports.
type ModelCapabilities struct {
	Temperature bool    `json:"temperature"`
	Reasoning   bool    `json:"reasoning"`
	Attachment  bool    `json:"attachment"`
	ToolCall    bool    `json:"toolcall"`
	Input       ModelIO `json:"input"`
	Output      ModelIO `json:"output"`
	// Interleaved is either a bool or {"field": "..."}.
	Interleaved any `json:"interleaved,omitempty"`
}

// ModelCost holds per-million-token pricing.
type ModelCost struct {
	Input  float64    `json:"input"`
	Output float64    `json:"output"`
	Cache  TokenCache `json:"cache"`
	// Tiers and ExperimentalOver200K are kept loosely typed.
	Tiers                []map[string]any `json:"tiers,omitempty"`
	ExperimentalOver200K map[string]any   `json:"experimentalOver200K,omitempty"`
}

// ModelLimit holds context window limits.
type ModelLimit struct {
	Context float64 `json:"context"`
	Input   float64 `json:"input,omitempty"`
	Output  float64 `json:"output"`
}

// Model describes an available model.
type Model struct {
	ID           string                    `json:"id"`
	ProviderID   string                    `json:"providerID"`
	API          ModelAPI                  `json:"api"`
	Name         string                    `json:"name"`
	Family       string                    `json:"family,omitempty"`
	Capabilities ModelCapabilities         `json:"capabilities"`
	Cost         ModelCost                 `json:"cost"`
	Limit        ModelLimit                `json:"limit"`
	Status       string                    `json:"status"` // alpha | beta | deprecated | active
	Options      map[string]any            `json:"options"`
	Headers      map[string]string         `json:"headers,omitempty"`
	ReleaseDate  string                    `json:"release_date"`
	Variants     map[string]map[string]any `json:"variants,omitempty"`
}

// Provider is a configured model provider.
type Provider struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Source  string           `json:"source"` // env | config | custom | api
	Env     []string         `json:"env"`
	Key     string           `json:"key,omitempty"`
	Options map[string]any   `json:"options"`
	Models  map[string]Model `json:"models"`
}

// AgentModel identifies the model an agent is configured to use.
type AgentModel struct {
	ModelID    string `json:"modelID"`
	ProviderID string `json:"providerID"`
}

// Agent is an available agent.
type Agent struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Mode        string            `json:"mode"` // subagent | primary | all
	Native      bool              `json:"native,omitempty"`
	Hidden      bool              `json:"hidden,omitempty"`
	TopP        float64           `json:"topP,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
	Color       string            `json:"color,omitempty"`
	Permission  PermissionRuleset `json:"permission"`
	Model       *AgentModel       `json:"model,omitempty"`
	Variant     string            `json:"variant,omitempty"`
	Prompt      string            `json:"prompt,omitempty"`
	Options     map[string]any    `json:"options"`
	Steps       float64           `json:"steps,omitempty"`
}

// Command is a slash command available on the server.
type Command struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Agent       string   `json:"agent,omitempty"`
	Model       string   `json:"model,omitempty"`
	Source      string   `json:"source,omitempty"` // command | mcp | skill
	Template    string   `json:"template"`
	Subtask     bool     `json:"subtask,omitempty"`
	Hints       []string `json:"hints"`
}

// Path reports the server's important directories.
type Path struct {
	Home      string `json:"home"`
	State     string `json:"state"`
	Config    string `json:"config"`
	Worktree  string `json:"worktree"`
	Directory string `json:"directory"`
}

// ProjectIcon holds display metadata for a project.
type ProjectIcon struct {
	URL      string `json:"url,omitempty"`
	Override string `json:"override,omitempty"`
	Color    string `json:"color,omitempty"`
}

// ProjectCommands holds project lifecycle commands.
type ProjectCommands struct {
	Start string `json:"start,omitempty"`
}

// ProjectTime holds project timestamps (unix milliseconds).
type ProjectTime struct {
	Created     int64 `json:"created"`
	Updated     int64 `json:"updated"`
	Initialized int64 `json:"initialized,omitempty"`
}

// Project is a project known to the server.
type Project struct {
	ID        string           `json:"id"`
	Worktree  string           `json:"worktree"`
	VCS       string           `json:"vcs,omitempty"` // "git"
	Name      string           `json:"name,omitempty"`
	Icon      *ProjectIcon     `json:"icon,omitempty"`
	Commands  *ProjectCommands `json:"commands,omitempty"`
	Time      ProjectTime      `json:"time"`
	Sandboxes []string         `json:"sandboxes"`
}

// VcsInfo reports version control information for the current project.
type VcsInfo struct {
	Branch        string `json:"branch,omitempty"`
	DefaultBranch string `json:"default_branch,omitempty"`
}

// File reports the status of a tracked file.
type File struct {
	Path    string `json:"path"`
	Added   int    `json:"added"`
	Removed int    `json:"removed"`
	Status  string `json:"status"` // added | deleted | modified
}

// FileNode is a directory listing entry.
type FileNode struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Absolute string `json:"absolute"`
	Type     string `json:"type"` // file | directory
	Ignored  bool   `json:"ignored"`
}

// FileHunk is a hunk in a structured patch.
type FileHunk struct {
	OldStart int      `json:"oldStart"`
	OldLines int      `json:"oldLines"`
	NewStart int      `json:"newStart"`
	NewLines int      `json:"newLines"`
	Lines    []string `json:"lines"`
}

// FilePatch is a structured diff of a file.
type FilePatch struct {
	OldFileName string     `json:"oldFileName"`
	NewFileName string     `json:"newFileName"`
	OldHeader   string     `json:"oldHeader,omitempty"`
	NewHeader   string     `json:"newHeader,omitempty"`
	Hunks       []FileHunk `json:"hunks"`
	Index       string     `json:"index,omitempty"`
}

// FileContent is the content of a file read from the server.
type FileContent struct {
	Type     string     `json:"type"` // text | binary
	Content  string     `json:"content"`
	Diff     string     `json:"diff,omitempty"`
	Patch    *FilePatch `json:"patch,omitempty"`
	Encoding string     `json:"encoding,omitempty"` // "base64" for binary content
	MimeType string     `json:"mimeType,omitempty"`
}

// Position is a zero-based line/character position.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is a text range.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// SymbolLocation is where a symbol is defined.
type SymbolLocation struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Symbol is a workspace symbol search result.
type Symbol struct {
	Name     string         `json:"name"`
	Kind     int            `json:"kind"`
	Location SymbolLocation `json:"location"`
}

// Todo is an entry in a session's task list.
type Todo struct {
	Content  string `json:"content"`
	Status   string `json:"status"`   // pending | in_progress | completed | cancelled
	Priority string `json:"priority"` // high | medium | low
}

// LSPStatus reports the state of a language server.
type LSPStatus struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Root   string `json:"root"`
	Status string `json:"status"` // connected | error
}

// FormatterStatus reports the state of a formatter.
type FormatterStatus struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
	Enabled    bool     `json:"enabled"`
}

// Health is the server health response.
type Health struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

// MatchText wraps a text value in ripgrep-style search results.
type MatchText struct {
	Text string `json:"text"`
}

// Submatch is a match within a matched line.
type Submatch struct {
	Match MatchText `json:"match"`
	Start int       `json:"start"`
	End   int       `json:"end"`
}

// Match is a text search result (ripgrep format).
type Match struct {
	Path           MatchText  `json:"path"`
	Lines          MatchText  `json:"lines"`
	LineNumber     int        `json:"line_number"`
	AbsoluteOffset int        `json:"absolute_offset"`
	Submatches     []Submatch `json:"submatches"`
}

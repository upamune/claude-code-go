package claude

import "encoding/json"

// PermissionMode represents how the SDK handles permissions for tool use
type PermissionMode string

// PermissionMode constants define how Claude Code handles permissions for tool use
const (
	// PermissionDefault uses the default permission mode
	PermissionDefault PermissionMode = "default"
	// PermissionAcceptEdits automatically accepts file edits
	PermissionAcceptEdits PermissionMode = "acceptEdits"
	// PermissionBypassPermissions bypasses all permission prompts
	PermissionBypassPermissions PermissionMode = "bypassPermissions"
	// PermissionPlan uses plan mode
	PermissionPlan PermissionMode = "plan"
)

// Options configures the behavior of Claude Code SDK
type Options struct {
	// Tools configuration
	AllowedTools    []string
	DisallowedTools []string

	// System prompt configuration
	CustomSystemPrompt string
	AppendSystemPrompt string

	// Working directory for the Claude Code CLI
	WorkingDir string

	// Token and turn limits
	MaxThinkingTokens *int
	MaxTurns          *int

	// MCP server configuration
	MCPServers map[string]MCPServerConfig

	// Path to the Claude Code CLI executable
	PathToClaudeCodeExecutable string // Default: "claude"

	// Permission handling
	PermissionMode           PermissionMode
	PermissionPromptToolName string

	// Session continuation
	Continue bool
	Resume   string

	// Model configuration
	Model         string
	FallbackModel string
}

// MCPServerConfig is an interface for MCP server configurations
type MCPServerConfig interface {
	mcpServerConfig() // Private method to seal the interface
	ToArg() string
}

// MCPStdioServerConfig represents a stdio-based MCP server
type MCPStdioServerConfig struct {
	Command string
	Args    []string
	Env     map[string]string
}

func (MCPStdioServerConfig) mcpServerConfig() {}

// ToArg converts the config to a CLI argument string
func (c MCPStdioServerConfig) ToArg() string {
	// Return struct as is for JSON marshaling
	b, _ := json.Marshal(c)
	return string(b)
}

// MCPSSEServerConfig represents an SSE-based MCP server
type MCPSSEServerConfig struct {
	URL     string
	Headers map[string]string
}

func (MCPSSEServerConfig) mcpServerConfig() {}

// ToArg converts the config to a CLI argument string
func (c MCPSSEServerConfig) ToArg() string {
	b, _ := json.Marshal(c)
	return string(b)
}

// MCPHTTPServerConfig represents an HTTP-based MCP server
type MCPHTTPServerConfig struct {
	URL     string
	Headers map[string]string
}

func (MCPHTTPServerConfig) mcpServerConfig() {}

// ToArg converts the config to a CLI argument string
func (c MCPHTTPServerConfig) ToArg() string {
	b, _ := json.Marshal(c)
	return string(b)
}

// Message is an interface for all message types from Claude Code CLI
type Message interface {
	messageType() string
}

// UserMessage represents a message from the user
type UserMessage struct {
	Type            string          `json:"type"`
	Message         json.RawMessage `json:"message"`
	ParentToolUseID *string         `json:"parent_tool_use_id"`
	SessionID       string          `json:"session_id"`
}

func (UserMessage) messageType() string { return "user" }

// AssistantMessage represents a message from the assistant
type AssistantMessage struct {
	Type            string          `json:"type"`
	Message         json.RawMessage `json:"message"`
	ParentToolUseID *string         `json:"parent_tool_use_id"`
	SessionID       string          `json:"session_id"`
}

func (AssistantMessage) messageType() string { return "assistant" }

// ResultMessage represents the final result of a Claude Code session
type ResultMessage struct {
	Type          string  `json:"type"`
	Subtype       string  `json:"subtype"`
	DurationMS    int64   `json:"duration_ms"`
	DurationAPIMS int64   `json:"duration_api_ms"`
	IsError       bool    `json:"is_error"`
	NumTurns      int     `json:"num_turns"`
	Result        string  `json:"result,omitempty"`
	SessionID     string  `json:"session_id"`
	TotalCostUSD  float64 `json:"total_cost_usd"`
	Usage         Usage   `json:"usage"`
}

func (ResultMessage) messageType() string { return "result" }

// SystemMessage represents system information from Claude Code CLI
type SystemMessage struct {
	Type           string            `json:"type"`
	Subtype        string            `json:"subtype"`
	APIKeySource   string            `json:"apiKeySource"`
	CWD            string            `json:"cwd"`
	SessionID      string            `json:"session_id"`
	Tools          []string          `json:"tools"`
	MCPServers     []MCPServerStatus `json:"mcp_servers"`
	Model          string            `json:"model"`
	PermissionMode string            `json:"permissionMode"`
}

func (SystemMessage) messageType() string { return "system" }

// PermissionRequestMessage represents a permission request for tool use
type PermissionRequestMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	Subtype   string `json:"subtype"`
}

func (PermissionRequestMessage) messageType() string { return "permission_request" }

// Usage represents token usage information
type Usage struct {
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}

// MCPServerStatus represents the status of an MCP server
type MCPServerStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

# Claude Code Go SDK

A Go SDK for integrating Claude AI into your Go applications through the Claude Code CLI. This SDK provides a type-safe, streaming-capable interface with comprehensive error handling.

## Installation

```bash
go get github.com/upamune/claude-code-go
```

## Prerequisites

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) installed and configured
- Valid Claude API key configured in the CLI

## Quick Start

### Basic Query

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    claude "github.com/upamune/claude-code-go"
)

func main() {
    ctx := context.Background()
    
    // Check if Claude CLI is available
    if !claude.IsClaudeAvailable() {
        log.Fatal("Claude Code CLI is not installed")
    }
    
    opts := &claude.Options{
        Model: "claude-3-5-sonnet-20241022",
        WorkingDir: "/path/to/project",
    }
    
    result, err := claude.Query(ctx, "Help me write a function", opts)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %s\n", result.Result)
    fmt.Printf("Cost: $%.4f\n", result.TotalCostUSD)
}
```

### Streaming Output

```go
stream, err := claude.QueryStream(ctx, "Write a detailed explanation", opts)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for msg := range stream.Messages {
    if msg.Err != nil {
        log.Printf("Error: %v\n", msg.Err)
        continue
    }
    
    switch m := msg.Message.(type) {
    case *claude.UserMessage:
        fmt.Printf("[USER] %s\n", m.Content)
    case *claude.AssistantMessage:
        if m.ToolCalls != nil {
            fmt.Printf("[ASSISTANT] Using tool: %s\n", m.ToolCalls[0].Name)
        } else {
            fmt.Printf("[ASSISTANT] %s\n", m.PartialContent)
        }
    case *claude.SystemMessage:
        fmt.Printf("[SYSTEM] %s: %s\n", m.Subtype, m.Subtype)
    case *claude.ResultMessage:
        fmt.Printf("\n=== Final Result ===\n%s\n", m.Result)
        fmt.Printf("Total cost: $%.4f\n", m.TotalCostUSD)
    }
}
```

### Using Custom Client

```go
// Create a custom client with specific executor
client := claude.NewClient()

// Or with custom command executor for testing
executor := &MyCustomExecutor{}
client := claude.NewClientWithExecutor(executor)

// Use the client
result, err := client.Query(ctx, "Your prompt", opts)
```

## Features

- 🚀 Simple, idiomatic Go interface
- 🔄 Real-time streaming support with channel-based API
- 🛡️ Type-safe message handling
- ⚡ Zero dependencies (uses only standard library)
- 🔧 Customizable client with dependency injection
- 📊 Built-in cost tracking
- 🎯 Comprehensive error types

## Configuration Options

```go
type Options struct {
    // Model selection
    Model         string   // e.g., "claude-3-5-sonnet-20241022"
    FallbackModel string   // Fallback if primary model unavailable
    
    // Working directory
    WorkingDir string      // Project directory for context
    
    // Tool restrictions
    AllowedTools    []string  // Whitelist specific tools
    DisallowedTools []string  // Blacklist specific tools
    
    // System prompts
    CustomSystemPrompt  string  // Replace default system prompt
    AppendSystemPrompt string  // Append to system prompt
    
    // Resource limits
    MaxThinkingTokens *int     // Limit thinking tokens
    MaxTurns         *int     // Limit conversation turns
    
    // MCP server configuration
    MCPServers map[string]MCPServerConfig  // MCP server configs
    
    // Permission handling
    PermissionMode           PermissionMode  // ask, allow, deny
    PermissionPromptToolName string          // Custom permission tool
    
    // Session management
    Continue bool    // Continue previous session
    Resume   string  // Resume specific session ID
    
    // Advanced
    PathToClaudeCodeExecutable string  // Custom CLI path
}
```

### MCP Server Types

```go
// Stdio MCP server
&StdioMCPServerConfig{
    Command: "node",
    Args:    []string{"server.js"},
    Env:     map[string]string{"NODE_ENV": "production"},
}

// SSE MCP server
&SSEMCPServerConfig{
    URL:     "http://localhost:8080/sse",
    Headers: map[string]string{"Authorization": "Bearer token"},
}

// HTTP MCP server
&HTTPMCPServerConfig{
    URL:     "https://api.example.com",
    Headers: map[string]string{"API-Key": "key"},
}
```

## Message Types

| Message Type | Description | Key Fields |
|-------------|-------------|------------|
| `UserMessage` | User input | `Content`, `Role` |
| `AssistantMessage` | Claude's responses | `Content`, `PartialContent`, `ToolCalls` |
| `ResultMessage` | Final session result | `Result`, `TotalCostUSD`, `CacheWriteTokens` |
| `SystemMessage` | System events | `Subtype` (info, warning, error) |
| `PermissionRequestMessage` | Tool permission requests | `ToolName`, `Arguments`, `Reason` |

## Error Handling

```go
result, err := claude.Query(ctx, prompt, opts)
if err != nil {
    switch e := err.(type) {
    case *claude.AbortError:
        // User cancelled operation
        fmt.Println("Operation cancelled")
    case *claude.ProcessError:
        // CLI process error
        fmt.Printf("Process error (exit %d): %s\n", e.ExitCode, e.Message)
    case *claude.ParseError:
        // JSON parsing error
        fmt.Printf("Parse error: %s\n", e.Message)
    case *claude.ConfigError:
        // Configuration error
        fmt.Printf("Config error in %s: %s\n", e.Field, e.Message)
    default:
        // Other errors
        log.Fatal(err)
    }
}
```

## Advanced Usage

### Permission Handling

```go
opts := &claude.Options{
    PermissionMode: claude.PermissionModeAsk,
    PermissionPromptToolName: "my_permission_handler",
}

stream, _ := claude.QueryStream(ctx, prompt, opts)
for msg := range stream.Messages {
    if perm, ok := msg.Message.(*claude.PermissionRequestMessage); ok {
        fmt.Printf("Permission requested for %s\n", perm.ToolName)
        // Handle permission request
    }
}
```

### Session Continuation

```go
// Continue previous session
opts := &claude.Options{
    Continue: true,
}

// Resume specific session
opts := &claude.Options{
    Resume: "session-id-123",
}
```

### Tool Restrictions

```go
// Only allow specific tools
opts := &claude.Options{
    AllowedTools: []string{"read_file", "write_file"},
}

// Disallow dangerous tools
opts := &claude.Options{
    DisallowedTools: []string{"run_command", "delete_file"},
}
```

## Examples

- [Basic Usage](examples/basic) - Simple query execution
- More examples coming soon...

## API Reference

### Package Functions

```go
// Check if Claude CLI is available
func IsClaudeAvailable() bool

// Execute a query (uses default client)
func Query(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error)

// Stream query results (uses default client)
func QueryStream(ctx context.Context, prompt string, opts *Options) (*MessageStream, error)

// Execute raw CLI command
func Exec(ctx context.Context, args []string) (*bytes.Buffer, error)
```

### Client Interface

```go
type Client interface {
    Query(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error)
    QueryStream(ctx context.Context, prompt string, opts *Options) (*MessageStream, error)
}

// Create default client
func NewClient() Client

// Create client with custom executor
func NewClientWithExecutor(executor CommandExecutor) Client
```

### MessageStream

```go
type MessageStream struct {
    Messages <-chan MessageOrError
}

// Close the stream
func (s *MessageStream) Close()
```

## Testing

The SDK is designed with testability in mind. You can inject custom command executors:

```go
type MockExecutor struct{}

func (m *MockExecutor) Execute(ctx context.Context, executable string, args []string, input string) ([]byte, error) {
    // Return mock response
    return []byte(`{"result": "mock result", "totalCostUsd": 0.01}`), nil
}

client := claude.NewClientWithExecutor(&MockExecutor{})
```

## Requirements

- Go 1.18 or higher
- Claude Code CLI installed and configured

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details

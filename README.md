# Claude Code Go SDK

A Go SDK for interacting with Claude Code programmatically. This SDK provides a Go interface to the Claude Code CLI, allowing you to integrate Claude's AI capabilities into your Go applications.

## Installation

```bash
go get github.com/upamune/claude-code-go
```

## Prerequisites

- Claude Code CLI installed and configured
- Valid Claude API key configured

## Usage

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

For real-time streaming output:

```go
err := claude.QueryStream(ctx, "Write a story", opts, func(msg claude.Message) error {
    switch m := msg.(type) {
    case *claude.SystemMessage:
        fmt.Printf("[SYSTEM] %s\n", m.Subtype)
    case *claude.AssistantMessage:
        fmt.Printf("[ASSISTANT] Processing...\n")
    case *claude.ResultMessage:
        fmt.Printf("[RESULT] %s\n", m.Result)
    }
    return nil
})
```

### Raw Command Execution

Execute raw claude commands:

```go
output, err := claude.Exec(ctx, []string{"--version"})
if err != nil {
    log.Fatal(err)
}
fmt.Println(output.String())
```

## Configuration Options

```go
type Options struct {
    // Tool restrictions
    AllowedTools    []string
    DisallowedTools []string
    
    // System prompts
    CustomSystemPrompt  string
    AppendSystemPrompt string
    
    // Working directory
    WorkingDir string
    
    // Token and turn limits
    MaxThinkingTokens *int
    MaxTurns         *int
    
    // MCP server configuration
    MCPServers map[string]MCPServerConfig
    
    // Path to Claude Code CLI (default: "claude")
    PathToClaudeCodeExecutable string
    
    // Permission handling
    PermissionMode           PermissionMode
    PermissionPromptToolName string
    
    // Session continuation
    Continue bool
    Resume   string
    
    // Model selection
    Model         string
    FallbackModel string
}
```

## Message Types

The SDK supports several message types:

- `UserMessage`: Messages from the user
- `AssistantMessage`: Messages from Claude
- `ResultMessage`: Final result of the session
- `SystemMessage`: System information and status
- `PermissionRequestMessage`: Permission requests for tool usage

## Error Handling

The SDK provides specific error types:

- `AbortError`: Operation was aborted
- `ProcessError`: Claude Code CLI process error
- `ParseError`: Message parsing error
- `ConfigError`: Configuration error

## Examples

See the [examples/basic](examples/basic) directory for a complete working example.

## API Methods

### Query
```go
func Query(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error)
```
Executes a query and returns the final result synchronously.

### QueryStream
```go
func QueryStream(ctx context.Context, prompt string, opts *Options, handler func(Message) error) error
```
Executes a query with streaming output, calling the handler for each message.

### Exec
```go
func Exec(ctx context.Context, args []string) (*bytes.Buffer, error)
```
Executes raw claude command with custom arguments.

## License

MIT

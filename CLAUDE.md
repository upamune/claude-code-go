# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go SDK for the Claude Code CLI, providing a programmatic interface to interact with Claude from Go applications. The SDK is designed with zero dependencies (using only the standard library) and offers both synchronous and streaming APIs.

## Development Commands

### Build and Test
```bash
# Install development tools (uses aqua)
task install

# Run all checks (format, lint, test)
task check

# Individual commands
task fmt    # Format code with goimports
task lint   # Run revive linter
task test   # Run tests
task build  # Build the project
```

### Development Tools
- **goimports**: Code formatter (v0.34.0)
- **revive**: Go linter (v1.10.0)
- **task**: Task runner (v3.44.0)
- All tools are managed via aqua (see aqua.yaml)

## Architecture

### Core Components

1. **Client Interface** (`client.go`): Main entry point for SDK users
   - `Client` interface with `Query` and `QueryStream` methods
   - Default implementation using OS command execution
   - Supports custom `CommandExecutor` injection for testing

2. **Message System** (`types.go`, `message.go`):
   - Typed message system with interfaces for different message types
   - JSON parsing with strict validation
   - Support for streaming and partial messages

3. **Command Building** (`builder.go`):
   - Converts Go `Options` struct to CLI arguments
   - Handles all Claude CLI flags and options
   - Special handling for MCP server configurations

4. **Error Handling** (`errors.go`):
   - Typed errors: `AbortError`, `ProcessError`, `ParseError`, `ConfigError`
   - Each error type provides specific context for debugging

### Key Design Patterns

- **Dependency Injection**: `CommandExecutor` interface allows for easy testing
- **Channel-based Streaming**: Uses Go channels for real-time message streaming
- **Builder Pattern**: `argumentsBuilder` constructs CLI arguments from options
- **Interface-based Design**: Core functionality exposed through interfaces

## Testing Strategy

- Unit tests for all core components (`*_test.go` files)
- Mock `CommandExecutor` for isolated testing
- Test utilities in `test_utils.go` for common test scenarios
- Examples in `examples/` directory demonstrate real usage

## Important Implementation Notes

1. **JSON Parsing**: The SDK expects JSONL format from Claude CLI output. Each line is parsed independently as a complete JSON object.

2. **Streaming**: The `QueryStream` method uses goroutines to read stdout/stderr concurrently, parsing messages in real-time.

3. **Process Management**: Proper cleanup is crucial - always defer `Close()` on message streams to prevent goroutine leaks.

4. **Error Propagation**: Errors from the Claude CLI are wrapped in typed errors for better handling by SDK users.

## Common Development Tasks

### Adding New Options
1. Add field to `Options` struct in `types.go`
2. Update `argumentsBuilder.build()` in `builder.go` to handle the new option
3. Add corresponding test in `builder_test.go`

### Adding New Message Types
1. Define the message struct in `types.go`
2. Implement the `Message` interface
3. Update `parseMessage()` in `message.go` to handle the new type
4. Add parsing test in `message_test.go`

### Debugging Tips
- Enable verbose output by checking stderr from the Claude process
- Use the `executor_test.go` mock executor to simulate Claude CLI responses
- Check `builder_test.go` for examples of how options translate to CLI arguments

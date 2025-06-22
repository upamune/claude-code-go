package claude

import "fmt"

// AbortError represents an operation that was aborted
type AbortError struct {
	Message string
}

func (e *AbortError) Error() string {
	if e.Message == "" {
		return "operation aborted"
	}
	return e.Message
}

// ProcessError represents an error from the Claude Code CLI process
type ProcessError struct {
	ExitCode int
	Message  string
}

func (e *ProcessError) Error() string {
	return fmt.Sprintf("process exited with code %d: %s", e.ExitCode, e.Message)
}

// ParseError represents an error parsing messages from the CLI
type ParseError struct {
	Line    string
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse message: %s (line: %s)", e.Message, e.Line)
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Value   string
	Reason  string
	Message string
}

func (e *ConfigError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("configuration error in field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("configuration error in field '%s' with value '%s': %s", e.Field, e.Value, e.Reason)
}

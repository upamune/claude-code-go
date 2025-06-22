package claude

import (
	"strings"
	"testing"
)

func TestAbortError(t *testing.T) {
	tests := []struct {
		name    string
		err     *AbortError
		wantMsg string
	}{
		{
			name:    "with message",
			err:     &AbortError{Message: "user cancelled operation"},
			wantMsg: "user cancelled operation",
		},
		{
			name:    "empty message",
			err:     &AbortError{Message: ""},
			wantMsg: "operation aborted",
		},
		{
			name:    "nil struct with empty message",
			err:     &AbortError{},
			wantMsg: "operation aborted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("AbortError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

func TestProcessError(t *testing.T) {
	tests := []struct {
		name    string
		err     *ProcessError
		wantMsg string
	}{
		{
			name: "exit code 1 with message",
			err: &ProcessError{
				ExitCode: 1,
				Message:  "command not found",
			},
			wantMsg: "process exited with code 1: command not found",
		},
		{
			name: "exit code 0 with message",
			err: &ProcessError{
				ExitCode: 0,
				Message:  "unexpected error",
			},
			wantMsg: "process exited with code 0: unexpected error",
		},
		{
			name: "negative exit code",
			err: &ProcessError{
				ExitCode: -1,
				Message:  "signal terminated",
			},
			wantMsg: "process exited with code -1: signal terminated",
		},
		{
			name: "empty message",
			err: &ProcessError{
				ExitCode: 127,
				Message:  "",
			},
			wantMsg: "process exited with code 127: ",
		},
		{
			name: "exit code 255",
			err: &ProcessError{
				ExitCode: 255,
				Message:  "fatal error",
			},
			wantMsg: "process exited with code 255: fatal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("ProcessError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name    string
		err     *ParseError
		wantMsg string
	}{
		{
			name: "with line and message",
			err: &ParseError{
				Line:    `{"invalid": json}`,
				Message: "unexpected character",
			},
			wantMsg: `failed to parse message: unexpected character (line: {"invalid": json})`,
		},
		{
			name: "empty line",
			err: &ParseError{
				Line:    "",
				Message: "empty input",
			},
			wantMsg: "failed to parse message: empty input (line: )",
		},
		{
			name: "long line",
			err: &ParseError{
				Line:    strings.Repeat("x", 100),
				Message: "too long",
			},
			wantMsg: "failed to parse message: too long (line: " + strings.Repeat("x", 100) + ")",
		},
		{
			name: "empty message",
			err: &ParseError{
				Line:    "some line",
				Message: "",
			},
			wantMsg: "failed to parse message:  (line: some line)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("ParseError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

func TestConfigError(t *testing.T) {
	tests := []struct {
		name    string
		err     *ConfigError
		wantMsg string
	}{
		{
			name: "with message only",
			err: &ConfigError{
				Field:   "Model",
				Message: "model is required",
			},
			wantMsg: "configuration error in field 'Model': model is required",
		},
		{
			name: "with value and reason",
			err: &ConfigError{
				Field:  "MaxTokens",
				Value:  "-10",
				Reason: "must be non-negative",
			},
			wantMsg: "configuration error in field 'MaxTokens' with value '-10': must be non-negative",
		},
		{
			name: "with all fields",
			err: &ConfigError{
				Field:   "Temperature",
				Value:   "3.5",
				Reason:  "must be between 0 and 2",
				Message: "invalid temperature", // Message takes precedence
			},
			wantMsg: "configuration error in field 'Temperature': invalid temperature",
		},
		{
			name: "empty message and reason",
			err: &ConfigError{
				Field: "APIKey",
				Value: "",
			},
			wantMsg: "configuration error in field 'APIKey' with value '': ",
		},
		{
			name: "special characters in value",
			err: &ConfigError{
				Field:  "Path",
				Value:  `/path/with/"quotes"`,
				Reason: "invalid characters",
			},
			wantMsg: `configuration error in field 'Path' with value '/path/with/"quotes"': invalid characters`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("ConfigError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that all error types implement the error interface
	var _ error = &AbortError{}
	var _ error = &ProcessError{}
	var _ error = &ParseError{}
	var _ error = &ConfigError{}

	// Test type assertions work correctly
	tests := []struct {
		name string
		err  error
		typ  string
	}{
		{"AbortError", &AbortError{}, "*claude.AbortError"},
		{"ProcessError", &ProcessError{}, "*claude.ProcessError"},
		{"ParseError", &ParseError{}, "*claude.ParseError"},
		{"ConfigError", &ConfigError{}, "*claude.ConfigError"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify error type
			switch tt.err.(type) {
			case *AbortError:
				if tt.typ != "*claude.AbortError" {
					t.Errorf("Expected AbortError, got %T", tt.err)
				}
			case *ProcessError:
				if tt.typ != "*claude.ProcessError" {
					t.Errorf("Expected ProcessError, got %T", tt.err)
				}
			case *ParseError:
				if tt.typ != "*claude.ParseError" {
					t.Errorf("Expected ParseError, got %T", tt.err)
				}
			case *ConfigError:
				if tt.typ != "*claude.ConfigError" {
					t.Errorf("Expected ConfigError, got %T", tt.err)
				}
			default:
				t.Errorf("Unknown error type: %T", tt.err)
			}
		})
	}
}

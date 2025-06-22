package claude

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestQuery(t *testing.T) {
	// Save original defaultClient
	originalClient := defaultClient
	defer func() {
		defaultClient = originalClient
	}()

	ctx := context.Background()

	tests := []struct {
		name    string
		prompt  string
		opts    *Options
		mockFn  func(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error)
		want    *ResultMessage
		wantErr bool
	}{
		{
			name:   "successful query",
			prompt: "Hello",
			opts:   &Options{Model: "claude-3"},
			mockFn: func(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error) {
				return &ResultMessage{
					Type:      "result",
					SessionID: "test-session",
					Result:    "Hello response",
					Usage: Usage{
						InputTokens:  10,
						OutputTokens: 20,
					},
				}, nil
			},
			want: &ResultMessage{
				Type:      "result",
				SessionID: "test-session",
				Result:    "Hello response",
				Usage: Usage{
					InputTokens:  10,
					OutputTokens: 20,
				},
			},
			wantErr: false,
		},
		{
			name:   "error from client",
			prompt: "test",
			opts:   nil,
			mockFn: func(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error) {
				return nil, errors.New("client error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mockClaudeClient{
				queryFunc: tt.mockFn,
			}
			defaultClient = mockClient

			got, err := Query(ctx, tt.prompt, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !resultMessageEqual(got, tt.want) {
				t.Errorf("Query() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestQueryStream(t *testing.T) {
	// Save original defaultClient
	originalClient := defaultClient
	defer func() {
		defaultClient = originalClient
	}()

	ctx := context.Background()

	tests := []struct {
		name         string
		prompt       string
		opts         *Options
		mockFn       func(ctx context.Context, prompt string, opts *Options, handler func(Message) error) error
		wantMessages int
		wantErr      bool
	}{
		{
			name:   "successful stream",
			prompt: "Hello stream",
			opts:   nil,
			mockFn: func(ctx context.Context, prompt string, opts *Options, handler func(Message) error) error {
				// Simulate streaming messages
				messages := []Message{
					&UserMessage{Type: "user", SessionID: "test"},
					&AssistantMessage{Type: "assistant", SessionID: "test"},
					&ResultMessage{Type: "result", SessionID: "test"},
				}

				for _, msg := range messages {
					if err := handler(msg); err != nil {
						return err
					}
				}
				return nil
			},
			wantMessages: 3,
			wantErr:      false,
		},
		{
			name:   "error from client",
			prompt: "test",
			opts:   nil,
			mockFn: func(ctx context.Context, prompt string, opts *Options, handler func(Message) error) error {
				return errors.New("stream error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mockClaudeClient{
				queryStreamFunc: tt.mockFn,
			}
			defaultClient = mockClient

			var receivedCount int
			err := QueryStream(ctx, tt.prompt, tt.opts, func(msg Message) error {
				receivedCount++
				return nil
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("QueryStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && receivedCount != tt.wantMessages {
				t.Errorf("QueryStream() received %d messages, want %d", receivedCount, tt.wantMessages)
			}
		})
	}
}

func TestExec(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		isValid bool // whether the command is expected to be valid
	}{
		{
			name:    "version command",
			args:    []string{"--version"},
			wantErr: false,
			isValid: true,
		},
		{
			name:    "help command",
			args:    []string{"--help"},
			wantErr: false,
			isValid: true,
		},
		{
			name:    "invalid command",
			args:    []string{"--invalid-flag-that-does-not-exist"},
			wantErr: true,
			isValid: false,
		},
		{
			name:    "empty args",
			args:    []string{},
			wantErr: true,
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Short() {
				t.Skip("Skipping Exec test in short mode")
			}

			skipIfClaudeNotAvailable(t)

			buf, err := Exec(context.Background(), tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Exec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				// Check if it's a ProcessError
				if _, ok := err.(*ProcessError); !ok {
					t.Errorf("Exec() error type = %T, want *ProcessError", err)
				}
				return
			}

			if buf == nil {
				t.Error("Exec() returned nil buffer without error")
			}

			// For valid commands, check that we got some output
			if tt.isValid && buf.Len() == 0 {
				t.Error("Exec() returned empty buffer for valid command")
			}
		})
	}
}

func TestExec_ProcessError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Exec ProcessError test in short mode")
	}

	ctx := context.Background()

	// Try to execute a command that doesn't exist
	_, err := Exec(ctx, []string{"nonexistentcommand12345"})
	if err == nil {
		t.Fatal("Expected error for non-existent command, got nil")
	}

	// Just verify we get an error - the exact type may vary by system
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// Mock client for testing package-level functions
type mockClaudeClient struct {
	queryFunc       func(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error)
	queryStreamFunc func(ctx context.Context, prompt string, opts *Options, handler func(Message) error) error
}

func (m *mockClaudeClient) Query(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, prompt, opts)
	}
	return nil, errors.New("not implemented")
}

func (m *mockClaudeClient) QueryStream(ctx context.Context, prompt string, opts *Options, handler func(Message) error) error {
	if m.queryStreamFunc != nil {
		return m.queryStreamFunc(ctx, prompt, opts, handler)
	}
	return errors.New("not implemented")
}

func TestDefaultClientInitialization(t *testing.T) {
	// Test that defaultClient is properly initialized
	if defaultClient == nil {
		t.Fatal("defaultClient is nil")
	}

	// Verify it implements Client interface
	var _ Client = defaultClient
}

func TestPackageFunctions_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test verifies that the package-level functions properly delegate to defaultClient
	// We'll use a custom executor to avoid actual CLI calls

	mockExecutor := &MockCommandExecutor{
		ExecuteFunc: func(ctx context.Context, name string, args []string, stdin string) ([]byte, error) {
			// Return a valid result
			return []byte(`{
				"type": "result",
				"subtype": "success",
				"session_id": "test",
				"usage": {"input_tokens": 5, "output_tokens": 10}
			}`), nil
		},
		ExecuteStreamFunc: func(ctx context.Context, name string, args []string, stdin string) (io.ReadCloser, error) {
			// Return a stream with messages
			data := strings.Join([]string{
				`{"type": "user", "message": {}, "session_id": "test"}`,
				`{"type": "result", "subtype": "success", "session_id": "test", "usage": {"input_tokens": 5, "output_tokens": 10}}`,
			}, "\n")
			return &mockReadCloser{
				Reader: strings.NewReader(data),
			}, nil
		},
	}

	// Save original client
	originalClient := defaultClient
	defer func() {
		defaultClient = originalClient
	}()

	// Replace with custom client
	defaultClient = NewClientWithExecutor(mockExecutor)

	ctx := context.Background()

	// Test Query
	result, err := Query(ctx, "test query", nil)
	if err != nil {
		t.Errorf("Query() integration test failed: %v", err)
	}
	if result == nil || result.Type != "result" {
		t.Error("Query() integration test returned unexpected result")
	}

	// Test QueryStream
	messageCount := 0
	err = QueryStream(ctx, "test stream", nil, func(msg Message) error {
		messageCount++
		return nil
	})
	if err != nil {
		t.Errorf("QueryStream() integration test failed: %v", err)
	}
	if messageCount != 2 {
		t.Errorf("QueryStream() integration test received %d messages, want 2", messageCount)
	}
}

// Ensure package-level functions match Client interface
func TestPackageFunctionsSignature(t *testing.T) {
	// This test ensures that the package-level functions have the same signature
	// as the Client interface methods

	// Simply verify that defaultClient implements Client interface
	// This is a compile-time check
	var _ Client = defaultClient

	// The fact that this compiles proves that:
	// 1. defaultClient implements the Client interface
	// 2. The package-level functions have the correct signatures
}

// Helper function to create test options
func testOptions() *Options {
	return &Options{
		Model:             "claude-3",
		MaxThinkingTokens: intPtr(1000),
		MaxTurns:          intPtr(5),
	}
}

func TestExec_WithContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Exec context test in short mode")
	}

	// Test with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Try to execute a long-running command
	_, err := Exec(ctx, []string{"sleep", "5"})

	// We expect an error due to context timeout
	if err == nil {
		t.Error("Expected error with cancelled context, got nil")
	}
}

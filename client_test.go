package claude

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

// MockMessageParser implements MessageParser for testing
type MockMessageParser struct {
	ParseMessageFunc func(line string) (Message, error)
}

func (m *MockMessageParser) ParseMessage(line string) (Message, error) {
	if m.ParseMessageFunc != nil {
		return m.ParseMessageFunc(line)
	}
	return ParseMessage(line)
}

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.executor == nil {
		t.Error("NewClient() executor is nil")
	}

	if client.builder == nil {
		t.Error("NewClient() builder is nil")
	}

	if client.parser == nil {
		t.Error("NewClient() parser is nil")
	}
}

func TestNewClientWithExecutor(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}
	client := NewClientWithExecutor(mockExecutor)

	if client == nil {
		t.Fatal("NewClientWithExecutor() returned nil")
	}

	if client.executor != mockExecutor {
		t.Error("NewClientWithExecutor() did not set custom executor")
	}

	if client.builder == nil {
		t.Error("NewClientWithExecutor() builder is nil")
	}

	if client.parser == nil {
		t.Error("NewClientWithExecutor() parser is nil")
	}
}

func TestClaudeClient_Query(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		prompt     string
		opts       *Options
		mockOutput []byte
		mockErr    error
		want       *ResultMessage
		wantErr    bool
		errType    error
	}{
		{
			name:   "successful query",
			prompt: "Hello, Claude!",
			opts:   &Options{Model: "claude-3"},
			mockOutput: []byte(`{
				"type": "result",
				"subtype": "success",
				"duration_ms": 1000,
				"duration_api_ms": 800,
				"is_error": false,
				"num_turns": 1,
				"result": "Hello! How can I help you?",
				"session_id": "test-session",
				"total_cost_usd": 0.001,
				"usage": {"input_tokens": 10, "output_tokens": 20}
			}`),
			want: &ResultMessage{
				Type:          "result",
				Subtype:       "success",
				DurationMS:    1000,
				DurationAPIMS: 800,
				IsError:       false,
				NumTurns:      1,
				Result:        "Hello! How can I help you?",
				SessionID:     "test-session",
				TotalCostUSD:  0.001,
				Usage: Usage{
					InputTokens:  10,
					OutputTokens: 20,
				},
			},
			wantErr: false,
		},
		{
			name:    "empty prompt",
			prompt:  "",
			opts:    nil,
			wantErr: true,
			errType: &ConfigError{},
		},
		{
			name:    "command execution error",
			prompt:  "test",
			opts:    nil,
			mockErr: errors.New("command failed"),
			wantErr: true,
		},
		{
			name:    "process error",
			prompt:  "test",
			opts:    nil,
			mockErr: &ProcessError{ExitCode: 1, Message: "CLI error"},
			wantErr: true,
			errType: &ProcessError{},
		},
		{
			name:       "invalid JSON response",
			prompt:     "test",
			opts:       nil,
			mockOutput: []byte(`{invalid json}`),
			wantErr:    true,
			errType:    &ParseError{},
		},
		{
			name:   "with custom options",
			prompt: "test",
			opts: &Options{
				Model:                      "claude-3-opus",
				PathToClaudeCodeExecutable: "/usr/local/bin/claude",
				MaxThinkingTokens:          intPtr(1000),
			},
			mockOutput: []byte(`{
				"type": "result",
				"subtype": "success",
				"is_error": false,
				"session_id": "test",
				"usage": {"input_tokens": 5, "output_tokens": 10}
			}`),
			want: &ResultMessage{
				Type:      "result",
				Subtype:   "success",
				IsError:   false,
				SessionID: "test",
				Usage: Usage{
					InputTokens:  5,
					OutputTokens: 10,
				},
			},
			wantErr: false,
		},
		{
			name:   "validation error",
			prompt: "test",
			opts: &Options{
				MaxThinkingTokens: intPtr(-1),
			},
			wantErr: true,
			errType: &ConfigError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockCommandExecutor{
				ExecuteFunc: func(ctx context.Context, name string, args []string, stdin string) ([]byte, error) {
					// Verify executable name
					expectedName := "claude"
					if tt.opts != nil && tt.opts.PathToClaudeCodeExecutable != "" {
						expectedName = tt.opts.PathToClaudeCodeExecutable
					}
					if name != expectedName {
						t.Errorf("Execute() name = %v, want %v", name, expectedName)
					}

					// Verify stdin contains prompt
					if stdin != tt.prompt {
						t.Errorf("Execute() stdin = %v, want %v", stdin, tt.prompt)
					}

					// Verify required args
					hasOutputFormat := false
					hasPrint := false
					for _, arg := range args {
						if arg == "--output-format" {
							hasOutputFormat = true
						}
						if arg == "--print" {
							hasPrint = true
						}
					}
					if !hasOutputFormat || !hasPrint {
						t.Error("Execute() missing required arguments")
					}

					return tt.mockOutput, tt.mockErr
				},
			}

			client := NewClientWithExecutor(mockExecutor)
			got, err := client.Query(ctx, tt.prompt, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errType != nil {
				// Check error type
				switch tt.errType.(type) {
				case *ConfigError:
					if _, ok := err.(*ConfigError); !ok {
						t.Errorf("Query() error type = %T, want *ConfigError", err)
					}
				case *ProcessError:
					if _, ok := err.(*ProcessError); !ok {
						t.Errorf("Query() error type = %T, want *ProcessError", err)
					}
				case *ParseError:
					if _, ok := err.(*ParseError); !ok {
						t.Errorf("Query() error type = %T, want *ParseError", err)
					}
				}
				return
			}

			if !tt.wantErr && !resultMessageEqual(got, tt.want) {
				t.Errorf("Query() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestClaudeClient_QueryStream(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		prompt            string
		opts              *Options
		streamData        []string
		streamErr         error
		wantMessages      []Message
		wantErr           bool
		handlerErr        error
		expectStreamError bool
	}{
		{
			name:   "successful stream",
			prompt: "Hello stream",
			opts:   nil,
			streamData: []string{
				`{"type": "user", "message": {"text": "Hello stream"}, "session_id": "test"}`,
				`{"type": "assistant", "message": {"text": "Hello!"}, "session_id": "test"}`,
				`{"type": "result", "subtype": "success", "session_id": "test", "usage": {"input_tokens": 5, "output_tokens": 10}}`,
			},
			wantMessages: []Message{
				&UserMessage{Type: "user", SessionID: "test"},
				&AssistantMessage{Type: "assistant", SessionID: "test"},
				&ResultMessage{Type: "result", Subtype: "success", SessionID: "test"},
			},
			wantErr: false,
		},
		{
			name:    "empty prompt",
			prompt:  "",
			opts:    nil,
			wantErr: true,
		},
		{
			name:      "stream error",
			prompt:    "test",
			opts:      nil,
			streamErr: errors.New("stream failed"),
			wantErr:   true,
		},
		{
			name:   "parse error in stream",
			prompt: "test",
			opts:   nil,
			streamData: []string{
				`{"type": "user", "message": {"text": "test"}, "session_id": "test"}`,
				`{invalid json}`,
			},
			wantMessages:      []Message{&UserMessage{Type: "user", SessionID: "test"}}, // One message before error
			wantErr:           false,                                                    // Stream starts successfully, error comes through channel
			expectStreamError: true,
		},
		{
			name:   "empty lines ignored",
			prompt: "test",
			opts:   nil,
			streamData: []string{
				``,
				`{"type": "user", "message": {"text": "test"}, "session_id": "test"}`,
				``,
				``,
				`{"type": "result", "subtype": "success", "session_id": "test", "usage": {"input_tokens": 5, "output_tokens": 10}}`,
			},
			wantMessages: []Message{
				&UserMessage{Type: "user", SessionID: "test"},
				&ResultMessage{Type: "result", Subtype: "success", SessionID: "test"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockCommandExecutor{
				ExecuteStreamFunc: func(ctx context.Context, name string, args []string, stdin string) (io.ReadCloser, error) {
					if tt.streamErr != nil {
						return nil, tt.streamErr
					}

					// Verify arguments contain stream-json
					hasStreamJSON := false
					for _, arg := range args {
						if arg == "stream-json" {
							hasStreamJSON = true
							break
						}
					}
					if !hasStreamJSON {
						t.Error("ExecuteStream() missing stream-json argument")
					}

					// Return mock stream data
					data := strings.Join(tt.streamData, "\n")
					return &mockReadCloser{
						Reader: strings.NewReader(data),
					}, nil
				},
			}

			client := NewClientWithExecutor(mockExecutor)

			stream, err := client.QueryStream(ctx, tt.prompt, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				var receivedMessages []Message
				var streamError error
				for msgOrErr := range stream.Messages {
					if msgOrErr.Err != nil {
						streamError = msgOrErr.Err
						break
					}
					receivedMessages = append(receivedMessages, msgOrErr.Message)
				}

				if tt.expectStreamError && streamError == nil {
					t.Error("Expected error in stream, got none")
				} else if !tt.expectStreamError && streamError != nil {
					t.Errorf("Unexpected error in stream: %v", streamError)
				}

				if len(receivedMessages) != len(tt.wantMessages) {
					t.Errorf("QueryStream() received %d messages, want %d",
						len(receivedMessages), len(tt.wantMessages))
					return
				}

				for i, msg := range receivedMessages {
					if msg.messageType() != tt.wantMessages[i].messageType() {
						t.Errorf("QueryStream() message[%d] type = %v, want %v",
							i, msg.messageType(), tt.wantMessages[i].messageType())
					}
				}
			}
		})
	}
}

func TestDefaultMessageParser(t *testing.T) {
	parser := &DefaultMessageParser{}

	tests := []struct {
		name    string
		line    string
		wantErr bool
	}{
		{
			name:    "valid user message",
			line:    `{"type": "user", "message": {"text": "hello"}, "session_id": "test"}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			line:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := parser.ParseMessage(tt.line)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && msg == nil {
				t.Error("ParseMessage() returned nil message without error")
			}
		})
	}
}

func TestClient_Interface(t *testing.T) {
	// Ensure ClaudeClient implements Client interface
	var _ Client = &ClaudeClient{}
}

func TestClaudeClient_QueryStream_ReaderError(t *testing.T) {
	ctx := context.Background()

	// Create a custom reader that simulates a stream with an error
	streamData := `{"type": "user", "message": {"text": "test"}, "session_id": "test"}` + "\n" +
		`{"type": "assistant", "message": {"text": "response"}, "session_id": "test"}` + "\n"

	mockExecutor := &MockCommandExecutor{
		ExecuteStreamFunc: func(ctx context.Context, name string, args []string, stdin string) (io.ReadCloser, error) {
			return &errorReaderCloser{
				errorAfterNReads: &errorAfterNReads{
					data:       streamData,
					errorAfter: 2, // Error after 2 messages
					err:        errors.New("simulated read error"),
				},
			}, nil
		},
	}

	client := NewClientWithExecutor(mockExecutor)

	stream, err := client.QueryStream(ctx, "test", nil)
	if err != nil {
		t.Fatalf("QueryStream() failed to start: %v", err)
	}

	var receivedCount int
	var streamErr error
	for msgOrErr := range stream.Messages {
		if msgOrErr.Err != nil {
			streamErr = msgOrErr.Err
			break
		}
		receivedCount++
	}

	// We expect to receive 2 messages before the error
	if streamErr == nil {
		t.Error("QueryStream() expected error, got nil")
	}

	if receivedCount != 2 {
		t.Errorf("QueryStream() received %d messages before error, want 2", receivedCount)
	}
}

// Helper types and functions

type errorReaderCloser struct {
	*errorAfterNReads
}

func (e *errorReaderCloser) Close() error {
	return nil
}

type errorAfterNReads struct {
	data       string
	errorAfter int
	err        error
	reader     *strings.Reader
	readCount  int
}

func (e *errorAfterNReads) Read(p []byte) (n int, err error) {
	if e.reader == nil {
		e.reader = strings.NewReader(e.data)
	}

	n, err = e.reader.Read(p)
	if err == nil && n > 0 {
		// Count lines read
		for i := 0; i < n; i++ {
			if p[i] == '\n' {
				e.readCount++
				if e.readCount >= e.errorAfter {
					return n, e.err
				}
			}
		}
	}
	return n, err
}

func intPtr(i int) *int {
	return &i
}

// Note: This function assumes both messages are of type *ResultMessage
// In a real test, you'd want a more comprehensive comparison
func resultMessageEqual(a, b *ResultMessage) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare all fields
	return a.Type == b.Type &&
		a.Subtype == b.Subtype &&
		a.DurationMS == b.DurationMS &&
		a.DurationAPIMS == b.DurationAPIMS &&
		a.IsError == b.IsError &&
		a.NumTurns == b.NumTurns &&
		a.Result == b.Result &&
		a.SessionID == b.SessionID &&
		a.TotalCostUSD == b.TotalCostUSD &&
		a.Usage.InputTokens == b.Usage.InputTokens &&
		a.Usage.OutputTokens == b.Usage.OutputTokens
}

func TestClaudeClient_BuildArgsIntegration(t *testing.T) {
	// Test that client properly integrates with ArgumentBuilder
	ctx := context.Background()

	calledWithArgs := false
	mockExecutor := &MockCommandExecutor{
		ExecuteFunc: func(ctx context.Context, name string, args []string, stdin string) ([]byte, error) {
			calledWithArgs = true

			// Check that options were properly converted to args
			hasModel := false
			for i, arg := range args {
				if arg == "--model" && i+1 < len(args) && args[i+1] == "claude-3-opus" {
					hasModel = true
				}
			}

			if !hasModel {
				t.Error("Execute() missing --model argument")
			}

			return []byte(`{"type": "result", "session_id": "test", "usage": {"input_tokens": 1, "output_tokens": 1}}`), nil
		},
	}

	client := NewClientWithExecutor(mockExecutor)
	opts := &Options{
		Model: "claude-3-opus",
	}

	_, err := client.Query(ctx, "test", opts)
	if err != nil {
		t.Errorf("Query() error = %v", err)
	}

	if !calledWithArgs {
		t.Error("Execute() was not called")
	}
}

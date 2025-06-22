package claude

import (
	"encoding/json"
	"testing"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    Message
		wantErr bool
		errType error
	}{
		{
			name: "valid user message",
			line: `{"type": "user", "message": {"text": "hello"}, "parent_tool_use_id": null, "session_id": "test-session"}`,
			want: &UserMessage{
				Type:            "user",
				Message:         json.RawMessage(`{"text": "hello"}`),
				ParentToolUseID: nil,
				SessionID:       "test-session",
			},
			wantErr: false,
		},
		{
			name: "valid user message with parent tool use id",
			line: `{"type": "user", "message": {"text": "hello"}, "parent_tool_use_id": "tool-123", "session_id": "test-session"}`,
			want: &UserMessage{
				Type:            "user",
				Message:         json.RawMessage(`{"text": "hello"}`),
				ParentToolUseID: stringPtr("tool-123"),
				SessionID:       "test-session",
			},
			wantErr: false,
		},
		{
			name: "valid assistant message",
			line: `{"type": "assistant", "message": {"text": "Hi there!"}, "parent_tool_use_id": null, "session_id": "test-session"}`,
			want: &AssistantMessage{
				Type:            "assistant",
				Message:         json.RawMessage(`{"text": "Hi there!"}`),
				ParentToolUseID: nil,
				SessionID:       "test-session",
			},
			wantErr: false,
		},
		{
			name: "valid result message",
			line: `{"type": "result", "subtype": "success", "duration_ms": 1000, "duration_api_ms": 800, "is_error": false, "num_turns": 2, "result": "Done", "session_id": "test-session", "total_cost_usd": 0.05, "usage": {"input_tokens": 100, "output_tokens": 50}}`,
			want: &ResultMessage{
				Type:          "result",
				Subtype:       "success",
				DurationMS:    1000,
				DurationAPIMS: 800,
				IsError:       false,
				NumTurns:      2,
				Result:        "Done",
				SessionID:     "test-session",
				TotalCostUSD:  0.05,
				Usage: Usage{
					InputTokens:  100,
					OutputTokens: 50,
				},
			},
			wantErr: false,
		},
		{
			name: "valid system message",
			line: `{"type": "system", "subtype": "info", "apiKeySource": "env", "cwd": "/home/user", "session_id": "test-session", "tools": ["bash", "read"], "mcp_servers": [], "model": "claude-3", "permissionMode": "ask"}`,
			want: &SystemMessage{
				Type:           "system",
				Subtype:        "info",
				APIKeySource:   "env",
				CWD:            "/home/user",
				SessionID:      "test-session",
				Tools:          []string{"bash", "read"},
				MCPServers:     []MCPServerStatus{},
				Model:          "claude-3",
				PermissionMode: "ask",
			},
			wantErr: false,
		},
		{
			name: "valid permission request message",
			line: `{"type": "permission_request", "session_id": "test-session", "subtype": "tool_use"}`,
			want: &PermissionRequestMessage{
				Type:      "permission_request",
				SessionID: "test-session",
				Subtype:   "tool_use",
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			line:    `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "empty line",
			line:    "",
			wantErr: true,
		},
		{
			name:    "missing type field",
			line:    `{"message": "hello"}`,
			wantErr: true,
		},
		{
			name:    "unknown message type",
			line:    `{"type": "unknown", "message": "hello"}`,
			wantErr: true,
		},
		{
			name: "malformed user message",
			line: `{"type": "user", "invalid_field": true}`,
			want: &UserMessage{
				Type:            "user",
				Message:         nil,
				ParentToolUseID: nil,
				SessionID:       "",
			},
			wantErr: false, // JSON parsing will succeed, but fields will be empty/default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMessage(tt.line)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				// Check if it's a ParseError
				if _, ok := err.(*ParseError); !ok {
					t.Errorf("ParseMessage() error type = %T, want *ParseError", err)
				}
				return
			}

			if !messageEqual(got, tt.want) {
				t.Errorf("ParseMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMessage_ParseError(t *testing.T) {
	line := `{invalid json}`
	_, err := ParseMessage(line)

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("Expected *ParseError, got %T", err)
	}

	if parseErr.Line != line {
		t.Errorf("ParseError.Line = %q, want %q", parseErr.Line, line)
	}

	if parseErr.Message == "" {
		t.Error("ParseError.Message should not be empty")
	}
}

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
		want string
	}{
		{"UserMessage", &UserMessage{}, "user"},
		{"AssistantMessage", &AssistantMessage{}, "assistant"},
		{"ResultMessage", &ResultMessage{}, "result"},
		{"SystemMessage", &SystemMessage{}, "system"},
		{"PermissionRequestMessage", &PermissionRequestMessage{}, "permission_request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.messageType(); got != tt.want {
				t.Errorf("%T.messageType() = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func messageEqual(a, b Message) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Type assertion and comparison for each message type
	switch va := a.(type) {
	case *UserMessage:
		vb, ok := b.(*UserMessage)
		if !ok {
			return false
		}
		return userMessageEqual(va, vb)
	case *AssistantMessage:
		vb, ok := b.(*AssistantMessage)
		if !ok {
			return false
		}
		return assistantMessageEqual(va, vb)
	case *ResultMessage:
		vb, ok := b.(*ResultMessage)
		if !ok {
			return false
		}
		return resultMessagesEqual(va, vb)
	case *SystemMessage:
		vb, ok := b.(*SystemMessage)
		if !ok {
			return false
		}
		return systemMessageEqual(va, vb)
	case *PermissionRequestMessage:
		vb, ok := b.(*PermissionRequestMessage)
		if !ok {
			return false
		}
		return permissionRequestMessageEqual(va, vb)
	default:
		return false
	}
}

func userMessageEqual(a, b *UserMessage) bool {
	if a.Type != b.Type || a.SessionID != b.SessionID {
		return false
	}
	if string(a.Message) != string(b.Message) {
		return false
	}
	if (a.ParentToolUseID == nil) != (b.ParentToolUseID == nil) {
		return false
	}
	if a.ParentToolUseID != nil && *a.ParentToolUseID != *b.ParentToolUseID {
		return false
	}
	return true
}

func assistantMessageEqual(a, b *AssistantMessage) bool {
	if a.Type != b.Type || a.SessionID != b.SessionID {
		return false
	}
	if string(a.Message) != string(b.Message) {
		return false
	}
	if (a.ParentToolUseID == nil) != (b.ParentToolUseID == nil) {
		return false
	}
	if a.ParentToolUseID != nil && *a.ParentToolUseID != *b.ParentToolUseID {
		return false
	}
	return true
}

func resultMessagesEqual(a, b *ResultMessage) bool {
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
		a.Usage.OutputTokens == b.Usage.OutputTokens &&
		a.Usage.CacheCreationInputTokens == b.Usage.CacheCreationInputTokens &&
		a.Usage.CacheReadInputTokens == b.Usage.CacheReadInputTokens
}

func systemMessageEqual(a, b *SystemMessage) bool {
	if a.Type != b.Type || a.Subtype != b.Subtype ||
		a.APIKeySource != b.APIKeySource || a.CWD != b.CWD ||
		a.SessionID != b.SessionID || a.Model != b.Model ||
		a.PermissionMode != b.PermissionMode {
		return false
	}

	if len(a.Tools) != len(b.Tools) {
		return false
	}
	for i := range a.Tools {
		if a.Tools[i] != b.Tools[i] {
			return false
		}
	}

	if len(a.MCPServers) != len(b.MCPServers) {
		return false
	}
	for i := range a.MCPServers {
		if a.MCPServers[i] != b.MCPServers[i] {
			return false
		}
	}

	return true
}

func permissionRequestMessageEqual(a, b *PermissionRequestMessage) bool {
	return a.Type == b.Type &&
		a.SessionID == b.SessionID &&
		a.Subtype == b.Subtype
}

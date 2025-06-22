package claude

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPermissionMode(t *testing.T) {
	tests := []struct {
		name string
		mode PermissionMode
		want string
	}{
		{"default", PermissionDefault, "default"},
		{"acceptEdits", PermissionAcceptEdits, "acceptEdits"},
		{"bypassPermissions", PermissionBypassPermissions, "bypassPermissions"},
		{"plan", PermissionPlan, "plan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.want {
				t.Errorf("PermissionMode = %v, want %v", tt.mode, tt.want)
			}
		})
	}
}

func TestMCPStdioServerConfig_ToArg(t *testing.T) {
	tests := []struct {
		name   string
		config MCPStdioServerConfig
		want   []string // expected JSON keys
	}{
		{
			name: "basic config",
			config: MCPStdioServerConfig{
				Command: "node",
				Args:    []string{"server.js"},
				Env:     map[string]string{"NODE_ENV": "production"},
			},
			want: []string{"Command", "node", "server.js", "NODE_ENV", "production"},
		},
		{
			name: "empty args and env",
			config: MCPStdioServerConfig{
				Command: "python",
				Args:    []string{},
				Env:     map[string]string{},
			},
			want: []string{"Command", "python"},
		},
		{
			name: "nil args and env",
			config: MCPStdioServerConfig{
				Command: "ruby",
			},
			want: []string{"Command", "ruby"},
		},
		{
			name: "multiple args",
			config: MCPStdioServerConfig{
				Command: "java",
				Args:    []string{"-jar", "server.jar", "--port", "8080"},
			},
			want: []string{"Command", "java", "-jar", "server.jar", "--port", "8080"},
		},
		{
			name: "multiple env vars",
			config: MCPStdioServerConfig{
				Command: "go",
				Args:    []string{"run", "main.go"},
				Env: map[string]string{
					"GOOS":   "linux",
					"GOARCH": "amd64",
					"PORT":   "3000",
				},
			},
			want: []string{"Command", "go", "run", "main.go", "GOOS", "GOARCH", "PORT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ToArg()

			// Verify it's valid JSON
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(got), &result); err != nil {
				t.Errorf("ToArg() returned invalid JSON: %v", err)
				return
			}

			// Check expected fields
			for _, expected := range tt.want {
				if !strings.Contains(got, expected) {
					t.Errorf("ToArg() = %v, want to contain %v", got, expected)
				}
			}
		})
	}
}

func TestMCPSSEServerConfig_ToArg(t *testing.T) {
	tests := []struct {
		name   string
		config MCPSSEServerConfig
		want   []string
	}{
		{
			name: "basic config",
			config: MCPSSEServerConfig{
				URL:     "https://example.com/sse",
				Headers: map[string]string{"Authorization": "Bearer token"},
			},
			want: []string{"URL", "https://example.com/sse", "Authorization", "Bearer token"},
		},
		{
			name: "no headers",
			config: MCPSSEServerConfig{
				URL:     "wss://stream.example.com",
				Headers: map[string]string{},
			},
			want: []string{"URL", "wss://stream.example.com"},
		},
		{
			name: "nil headers",
			config: MCPSSEServerConfig{
				URL: "https://events.example.com",
			},
			want: []string{"URL", "https://events.example.com"},
		},
		{
			name: "multiple headers",
			config: MCPSSEServerConfig{
				URL: "https://api.example.com/events",
				Headers: map[string]string{
					"X-API-Key":    "secret",
					"Content-Type": "text/event-stream",
					"User-Agent":   "Claude-SDK/1.0",
				},
			},
			want: []string{"URL", "https://api.example.com/events", "X-API-Key", "Content-Type", "User-Agent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ToArg()

			// Verify it's valid JSON
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(got), &result); err != nil {
				t.Errorf("ToArg() returned invalid JSON: %v", err)
				return
			}

			// Check expected fields
			for _, expected := range tt.want {
				if !strings.Contains(got, expected) {
					t.Errorf("ToArg() = %v, want to contain %v", got, expected)
				}
			}
		})
	}
}

func TestMCPHTTPServerConfig_ToArg(t *testing.T) {
	tests := []struct {
		name   string
		config MCPHTTPServerConfig
		want   []string
	}{
		{
			name: "basic config",
			config: MCPHTTPServerConfig{
				URL:     "https://api.example.com",
				Headers: map[string]string{"X-API-Key": "secret"},
			},
			want: []string{"URL", "https://api.example.com", "X-API-Key", "secret"},
		},
		{
			name: "no headers",
			config: MCPHTTPServerConfig{
				URL:     "http://localhost:8080",
				Headers: map[string]string{},
			},
			want: []string{"URL", "http://localhost:8080"},
		},
		{
			name: "complex headers",
			config: MCPHTTPServerConfig{
				URL: "https://api.service.com/v2",
				Headers: map[string]string{
					"Authorization": "Bearer very-long-token-string",
					"X-Request-ID":  "uuid-1234-5678-90ab",
					"Accept":        "application/json",
				},
			},
			want: []string{"URL", "https://api.service.com/v2", "Authorization", "X-Request-ID", "Accept"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ToArg()

			// Verify it's valid JSON
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(got), &result); err != nil {
				t.Errorf("ToArg() returned invalid JSON: %v", err)
				return
			}

			// Check expected fields
			for _, expected := range tt.want {
				if !strings.Contains(got, expected) {
					t.Errorf("ToArg() = %v, want to contain %v", got, expected)
				}
			}
		})
	}
}

func TestMCPServerConfig_Interface(t *testing.T) {
	// Test that all types implement the interface correctly
	var _ MCPServerConfig = &MCPStdioServerConfig{}
	var _ MCPServerConfig = &MCPSSEServerConfig{}
	var _ MCPServerConfig = &MCPHTTPServerConfig{}

	// Test interface methods
	configs := []MCPServerConfig{
		&MCPStdioServerConfig{Command: "test"},
		&MCPSSEServerConfig{URL: "test"},
		&MCPHTTPServerConfig{URL: "test"},
	}

	for i, config := range configs {
		// Should not panic
		config.mcpServerConfig()

		// ToArg should return valid JSON
		arg := config.ToArg()
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(arg), &result); err != nil {
			t.Errorf("Config[%d].ToArg() returned invalid JSON: %v", i, err)
		}
	}
}

func TestUsage(t *testing.T) {
	tests := []struct {
		name  string
		usage Usage
		want  map[string]int
	}{
		{
			name: "all fields set",
			usage: Usage{
				CacheCreationInputTokens: 100,
				CacheReadInputTokens:     200,
				InputTokens:              300,
				OutputTokens:             400,
			},
			want: map[string]int{
				"cache_creation": 100,
				"cache_read":     200,
				"input":          300,
				"output":         400,
			},
		},
		{
			name: "only required fields",
			usage: Usage{
				InputTokens:  100,
				OutputTokens: 50,
			},
			want: map[string]int{
				"input":  100,
				"output": 50,
			},
		},
		{
			name:  "zero values",
			usage: Usage{},
			want: map[string]int{
				"input":  0,
				"output": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON and back to verify omitempty behavior
			data, err := json.Marshal(tt.usage)
			if err != nil {
				t.Fatalf("Failed to marshal Usage: %v", err)
			}

			var result map[string]int
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal Usage: %v", err)
			}

			// Check expected fields
			if tt.usage.InputTokens != tt.want["input"] {
				t.Errorf("InputTokens = %v, want %v", tt.usage.InputTokens, tt.want["input"])
			}
			if tt.usage.OutputTokens != tt.want["output"] {
				t.Errorf("OutputTokens = %v, want %v", tt.usage.OutputTokens, tt.want["output"])
			}

			// Check cache fields only if they're set
			if tt.usage.CacheCreationInputTokens > 0 &&
				tt.usage.CacheCreationInputTokens != tt.want["cache_creation"] {
				t.Errorf("CacheCreationInputTokens = %v, want %v",
					tt.usage.CacheCreationInputTokens, tt.want["cache_creation"])
			}
			if tt.usage.CacheReadInputTokens > 0 &&
				tt.usage.CacheReadInputTokens != tt.want["cache_read"] {
				t.Errorf("CacheReadInputTokens = %v, want %v",
					tt.usage.CacheReadInputTokens, tt.want["cache_read"])
			}
		})
	}
}

func TestMCPServerStatus(t *testing.T) {
	tests := []struct {
		name   string
		status MCPServerStatus
	}{
		{
			name: "successful status",
			status: MCPServerStatus{
				Name:   "test-server",
				Status: "connected",
			},
		},
		{
			name: "error status",
			status: MCPServerStatus{
				Name:   "failed-server",
				Status: "error",
				Error:  "connection refused",
			},
		},
		{
			name: "empty error field omitted",
			status: MCPServerStatus{
				Name:   "running-server",
				Status: "running",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.status)
			if err != nil {
				t.Fatalf("Failed to marshal MCPServerStatus: %v", err)
			}

			// Check that error field is omitted when empty
			if tt.status.Error == "" && strings.Contains(string(data), "\"error\"") {
				t.Errorf("Expected error field to be omitted, but found in JSON: %s", data)
			}

			// Unmarshal back
			var result MCPServerStatus
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal MCPServerStatus: %v", err)
			}

			if result.Name != tt.status.Name {
				t.Errorf("Name = %v, want %v", result.Name, tt.status.Name)
			}
			if result.Status != tt.status.Status {
				t.Errorf("Status = %v, want %v", result.Status, tt.status.Status)
			}
			if result.Error != tt.status.Error {
				t.Errorf("Error = %v, want %v", result.Error, tt.status.Error)
			}
		})
	}
}

func TestOptions_Defaults(t *testing.T) {
	// Test that Options can be created with zero values
	opts := &Options{}

	// Verify no panic when accessing fields
	_ = opts.AllowedTools
	_ = opts.DisallowedTools
	_ = opts.CustomSystemPrompt
	_ = opts.AppendSystemPrompt
	_ = opts.WorkingDir
	_ = opts.MaxThinkingTokens
	_ = opts.MaxTurns
	_ = opts.MCPServers
	_ = opts.PathToClaudeCodeExecutable
	_ = opts.PermissionMode
	_ = opts.PermissionPromptToolName
	_ = opts.Continue
	_ = opts.Resume
	_ = opts.Model
	_ = opts.FallbackModel

	// Test that nil pointers are handled correctly
	if opts.MaxThinkingTokens != nil {
		t.Error("Expected MaxThinkingTokens to be nil")
	}
	if opts.MaxTurns != nil {
		t.Error("Expected MaxTurns to be nil")
	}
}

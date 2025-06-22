package claude

import (
	"reflect"
	"strings"
	"testing"
)

func TestArgumentBuilder_BuildArgs(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name string
		opts *Options
		want []string
	}{
		{
			name: "nil options",
			opts: nil,
			want: []string{},
		},
		{
			name: "empty options",
			opts: &Options{},
			want: []string{},
		},
		{
			name: "model configuration",
			opts: &Options{
				Model:         "claude-3-opus",
				FallbackModel: "claude-3-sonnet",
			},
			want: []string{
				"--model", "claude-3-opus",
				"--fallback-model", "claude-3-sonnet",
			},
		},
		{
			name: "session configuration",
			opts: &Options{
				Continue: true,
				Resume:   "session-123",
			},
			want: []string{
				"--continue",
				"--resume", "session-123",
			},
		},
		{
			name: "system prompt configuration",
			opts: &Options{
				CustomSystemPrompt: "You are a helpful assistant",
				AppendSystemPrompt: "Always be concise",
			},
			want: []string{
				"--system-prompt", "You are a helpful assistant",
				"--append-system-prompt", "Always be concise",
			},
		},
		{
			name: "tool configuration",
			opts: &Options{
				AllowedTools:    []string{"bash", "read", "write"},
				DisallowedTools: []string{"delete", "execute"},
			},
			want: []string{
				"--allowed-tools", "bash,read,write",
				"--disallowed-tools", "delete,execute",
			},
		},
		{
			name: "token and turn limits",
			opts: &Options{
				MaxThinkingTokens: intPtr(1000),
				MaxTurns:          intPtr(5),
			},
			want: []string{
				"--max-thinking-tokens", "1000",
				"--max-turns", "5",
			},
		},
		{
			name: "permission configuration",
			opts: &Options{
				PermissionMode:           PermissionDefault,
				PermissionPromptToolName: "dangerous-tool",
			},
			want: []string{
				"--permission-mode", "default",
				"--permission-prompt-tool-name", "dangerous-tool",
			},
		},
		{
			name: "MCP servers JSON",
			opts: &Options{
				MCPServers: map[string]MCPServerConfig{
					"server1": &MCPStdioServerConfig{
						Command: "node",
						Args:    []string{"server.js"},
						Env:     map[string]string{"NODE_ENV": "production"},
					},
				},
			},
			want: []string{
				"--mcp-servers", `{"server1":{"Command":"node","Args":["server.js"],"Env":{"NODE_ENV":"production"}}}`,
			},
		},
		{
			name: "multiple MCP servers",
			opts: &Options{
				MCPServers: map[string]MCPServerConfig{
					"stdio-server": &MCPStdioServerConfig{
						Command: "python",
						Args:    []string{"server.py"},
					},
					"sse-server": &MCPSSEServerConfig{
						URL:     "https://example.com/sse",
						Headers: map[string]string{"Authorization": "Bearer token"},
					},
					"http-server": &MCPHTTPServerConfig{
						URL:     "https://api.example.com",
						Headers: map[string]string{"X-API-Key": "secret"},
					},
				},
			},
			want: []string{
				"--mcp-servers",
			},
		},
		{
			name: "all options combined",
			opts: &Options{
				Model:                    "claude-3-opus",
				Continue:                 true,
				CustomSystemPrompt:       "Be helpful",
				AllowedTools:             []string{"read"},
				MaxThinkingTokens:        intPtr(500),
				PermissionMode:           PermissionPlan,
				PermissionPromptToolName: "write",
			},
			want: []string{
				"--model", "claude-3-opus",
				"--continue",
				"--system-prompt", "Be helpful",
				"--allowed-tools", "read",
				"--max-thinking-tokens", "500",
				"--permission-mode", "plan",
				"--permission-prompt-tool-name", "write",
			},
		},
		{
			name: "empty string values ignored",
			opts: &Options{
				Model:                    "",
				Resume:                   "",
				CustomSystemPrompt:       "",
				AppendSystemPrompt:       "",
				PermissionPromptToolName: "",
			},
			want: []string{},
		},
		{
			name: "zero values ignored",
			opts: &Options{
				MaxThinkingTokens: intPtr(0),
				MaxTurns:          intPtr(0),
			},
			want: []string{
				"--max-thinking-tokens", "0",
				"--max-turns", "0",
			},
		},
	}

	builder := &ArgumentBuilder{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := builder.BuildArgs(tt.opts)

			// For MCP servers test, we need to check if JSON is present
			if tt.name == "multiple MCP servers" {
				if len(got) != 2 || got[0] != "--mcp-servers" {
					t.Errorf("BuildArgs() = %v, want --mcp-servers with JSON", got)
					return
				}
				// Just verify it's valid JSON containing expected servers
				json := got[1]
				if !strings.Contains(json, "stdio-server") ||
					!strings.Contains(json, "sse-server") ||
					!strings.Contains(json, "http-server") {
					t.Errorf("BuildArgs() MCP JSON missing expected servers: %s", json)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArgumentBuilder_Validate(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name    string
		opts    *Options
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantErr: false,
		},
		{
			name: "valid options",
			opts: &Options{
				Model:             "claude-3",
				MaxThinkingTokens: intPtr(1000),
				MaxTurns:          intPtr(10),
				PermissionMode:    PermissionDefault,
			},
			wantErr: false,
		},
		{
			name: "negative MaxThinkingTokens",
			opts: &Options{
				MaxThinkingTokens: intPtr(-1),
			},
			wantErr: true,
			errMsg:  "MaxThinkingTokens",
		},
		{
			name: "negative MaxTurns",
			opts: &Options{
				MaxTurns: intPtr(-5),
			},
			wantErr: true,
			errMsg:  "MaxTurns",
		},
		{
			name: "invalid PermissionMode",
			opts: &Options{
				PermissionMode: "invalid",
			},
			wantErr: true,
			errMsg:  "PermissionMode",
		},
		{
			name: "nil MCP server",
			opts: &Options{
				MCPServers: map[string]MCPServerConfig{
					"server1": nil,
				},
			},
			wantErr: true,
			errMsg:  "MCPServers[server1]",
		},
		{
			name: "valid permission modes",
			opts: &Options{
				PermissionMode: PermissionDefault,
			},
			wantErr: false,
		},
		{
			name: "another valid permission mode",
			opts: &Options{
				PermissionMode: PermissionBypassPermissions,
			},
			wantErr: false,
		},
		{
			name: "empty permission mode is valid",
			opts: &Options{
				PermissionMode: "",
			},
			wantErr: false,
		},
		{
			name: "zero MaxThinkingTokens is valid",
			opts: &Options{
				MaxThinkingTokens: intPtr(0),
			},
			wantErr: false,
		},
		{
			name: "zero MaxTurns is valid",
			opts: &Options{
				MaxTurns: intPtr(0),
			},
			wantErr: false,
		},
		{
			name: "valid MCP servers",
			opts: &Options{
				MCPServers: map[string]MCPServerConfig{
					"server1": &MCPStdioServerConfig{Command: "node"},
					"server2": &MCPSSEServerConfig{URL: "https://example.com"},
					"server3": &MCPHTTPServerConfig{URL: "https://api.example.com"},
				},
			},
			wantErr: false,
		},
	}

	builder := &ArgumentBuilder{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := builder.Validate(tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				configErr, ok := err.(*ConfigError)
				if !ok {
					t.Errorf("Validate() error type = %T, want *ConfigError", err)
					return
				}

				if tt.errMsg != "" && !strings.Contains(configErr.Field, tt.errMsg) {
					t.Errorf("Validate() error field = %v, want to contain %v", configErr.Field, tt.errMsg)
				}
			}
		})
	}
}

func TestArgumentBuilder_BuildArgs_EdgeCases(t *testing.T) {
	builder := &ArgumentBuilder{}

	t.Run("empty slices", func(t *testing.T) {
		opts := &Options{
			AllowedTools:    []string{},
			DisallowedTools: []string{},
		}
		got := builder.BuildArgs(opts)
		if len(got) != 0 {
			t.Errorf("BuildArgs() with empty slices = %v, want empty", got)
		}
	})

	t.Run("single tool in slices", func(t *testing.T) {
		opts := &Options{
			AllowedTools:    []string{"bash"},
			DisallowedTools: []string{"rm"},
		}
		got := builder.BuildArgs(opts)
		want := []string{
			"--allowed-tools", "bash",
			"--disallowed-tools", "rm",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("BuildArgs() = %v, want %v", got, want)
		}
	})

	t.Run("tools with special characters", func(t *testing.T) {
		opts := &Options{
			AllowedTools: []string{"tool:special", "tool-with-dash", "tool_underscore"},
		}
		got := builder.BuildArgs(opts)
		want := []string{
			"--allowed-tools", "tool:special,tool-with-dash,tool_underscore",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("BuildArgs() = %v, want %v", got, want)
		}
	})
}

func TestArgumentBuilder_Validate_ConfigError(t *testing.T) {
	builder := &ArgumentBuilder{}
	intPtr := func(i int) *int { return &i }

	opts := &Options{
		MaxThinkingTokens: intPtr(-10),
	}

	err := builder.Validate(opts)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	configErr, ok := err.(*ConfigError)
	if !ok {
		t.Fatalf("Expected *ConfigError, got %T", err)
	}

	if configErr.Field != "MaxThinkingTokens" {
		t.Errorf("ConfigError.Field = %v, want MaxThinkingTokens", configErr.Field)
	}

	if configErr.Value != "-10" {
		t.Errorf("ConfigError.Value = %v, want -10", configErr.Value)
	}

	if configErr.Reason != "must be non-negative" {
		t.Errorf("ConfigError.Reason = %v, want 'must be non-negative'", configErr.Reason)
	}
}

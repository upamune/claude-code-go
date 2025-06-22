package claude

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ArgumentBuilder builds command line arguments for Claude CLI
type ArgumentBuilder struct{}

// BuildArgs constructs command line arguments from options
func (b *ArgumentBuilder) BuildArgs(opts *Options) []string {
	args := []string{}

	if opts == nil {
		return args
	}

	// Model configuration
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	if opts.FallbackModel != "" {
		args = append(args, "--fallback-model", opts.FallbackModel)
	}

	// Session configuration
	if opts.Continue {
		args = append(args, "--continue")
	}

	if opts.Resume != "" {
		args = append(args, "--resume", opts.Resume)
	}

	// System prompt configuration
	if opts.CustomSystemPrompt != "" {
		args = append(args, "--system-prompt", opts.CustomSystemPrompt)
	}

	if opts.AppendSystemPrompt != "" {
		args = append(args, "--append-system-prompt", opts.AppendSystemPrompt)
	}

	// Tool configuration
	if len(opts.AllowedTools) > 0 {
		args = append(args, "--allowed-tools", strings.Join(opts.AllowedTools, ","))
	}

	if len(opts.DisallowedTools) > 0 {
		args = append(args, "--disallowed-tools", strings.Join(opts.DisallowedTools, ","))
	}

	// Token and turn limits
	if opts.MaxThinkingTokens != nil {
		args = append(args, "--max-thinking-tokens", strconv.Itoa(*opts.MaxThinkingTokens))
	}

	if opts.MaxTurns != nil {
		args = append(args, "--max-turns", strconv.Itoa(*opts.MaxTurns))
	}

	// Permission configuration
	if opts.PermissionMode != "" {
		args = append(args, "--permission-mode", string(opts.PermissionMode))
	}

	if opts.PermissionPromptToolName != "" {
		args = append(args, "--permission-prompt-tool-name", opts.PermissionPromptToolName)
	}

	// MCP servers (JSON format)
	if len(opts.MCPServers) > 0 {
		mcpConfig, _ := json.Marshal(opts.MCPServers)
		args = append(args, "--mcp-servers", string(mcpConfig))
	}

	return args
}

// Validate checks if options are valid
func (b *ArgumentBuilder) Validate(opts *Options) error {
	if opts == nil {
		return nil
	}

	// Validate numeric constraints
	if opts.MaxThinkingTokens != nil && *opts.MaxThinkingTokens < 0 {
		return &ConfigError{Field: "MaxThinkingTokens", Value: strconv.Itoa(*opts.MaxThinkingTokens), Reason: "must be non-negative"}
	}

	if opts.MaxTurns != nil && *opts.MaxTurns < 0 {
		return &ConfigError{Field: "MaxTurns", Value: strconv.Itoa(*opts.MaxTurns), Reason: "must be non-negative"}
	}

	// Validate MCP server configs
	for name, server := range opts.MCPServers {
		if server == nil {
			return &ConfigError{Field: fmt.Sprintf("MCPServers[%s]", name), Value: "nil", Reason: "server config cannot be nil"}
		}
	}

	// Validate permission mode
	if opts.PermissionMode != "" {
		validModes := map[PermissionMode]bool{
			PermissionDefault:           true,
			PermissionAcceptEdits:       true,
			PermissionBypassPermissions: true,
			PermissionPlan:              true,
		}
		if !validModes[opts.PermissionMode] {
			return &ConfigError{Field: "PermissionMode", Value: string(opts.PermissionMode), Reason: "must be 'default', 'acceptEdits', 'bypassPermissions', or 'plan'"}
		}
	}

	return nil
}

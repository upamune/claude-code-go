// Package claude provides a Go SDK for interacting with Claude Code CLI.
// It allows you to programmatically execute Claude queries and process responses.
package claude

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// defaultClient is the package-level client instance
var defaultClient Client = NewClient()

// IsClaudeAvailable checks if the Claude CLI is available in the system PATH
func IsClaudeAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// Query executes a Claude Code query and returns the result
func Query(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error) {
	return defaultClient.Query(ctx, prompt, opts)
}

// QueryStream executes a Claude Code query and returns a channel of messages
func QueryStream(ctx context.Context, prompt string, opts *Options) (*MessageStream, error) {
	return defaultClient.QueryStream(ctx, prompt, opts)
}

// Exec executes a raw claude command with custom arguments
func Exec(ctx context.Context, args []string) (*bytes.Buffer, error) {
	cmd := exec.CommandContext(ctx, "claude", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, &ProcessError{
				ExitCode: exitErr.ExitCode(),
				Message:  string(exitErr.Stderr),
			}
		}
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}
	return bytes.NewBuffer(output), nil
}

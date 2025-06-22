package claude

import (
	"context"
	"encoding/json"
	"fmt"
)

// Client interface for Claude Code interaction
type Client interface {
	Query(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error)
	QueryStream(ctx context.Context, prompt string, opts *Options) (*MessageStream, error)
}

// Compile-time check that clientImpl implements Client interface
var _ Client = (*clientImpl)(nil)

// clientImpl implements the Client interface
type clientImpl struct {
	executor CommandExecutor
	builder  *ArgumentBuilder
	parser   MessageParser
}

// MessageParser interface for parsing messages
type MessageParser interface {
	ParseMessage(line string) (Message, error)
}

// DefaultMessageParser implements MessageParser
type DefaultMessageParser struct{}

// ParseMessage parses a JSON line into a Message
func (p *DefaultMessageParser) ParseMessage(line string) (Message, error) {
	return ParseMessage(line)
}

// NewClient creates a new Client with default components
func NewClient() Client {
	return &clientImpl{
		executor: &DefaultCommandExecutor{},
		builder:  &ArgumentBuilder{},
		parser:   &DefaultMessageParser{},
	}
}

// NewClientWithExecutor creates a new Client with a custom executor
func NewClientWithExecutor(executor CommandExecutor) Client {
	return &clientImpl{
		executor: executor,
		builder:  &ArgumentBuilder{},
		parser:   &DefaultMessageParser{},
	}
}

// Query executes a Claude Code query and returns the result
func (c *clientImpl) Query(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error) {
	if prompt == "" {
		return nil, &ConfigError{
			Field:   "prompt",
			Message: "prompt is required",
		}
	}

	if opts == nil {
		opts = &Options{}
	}

	// Validate options
	if err := c.builder.Validate(opts); err != nil {
		return nil, err
	}

	// Set defaults
	executable := opts.PathToClaudeCodeExecutable
	if executable == "" {
		executable = "claude"
	}

	// Build arguments
	args := append([]string{"--print", "--output-format", "json"}, c.builder.BuildArgs(opts)...)

	// Execute command
	output, err := c.executor.Execute(ctx, executable, args, prompt, opts.WorkingDir)
	if err != nil {
		if processErr, ok := err.(*ProcessError); ok {
			return nil, processErr
		}
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	// Parse JSON response
	var result ResultMessage
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, &ParseError{
			Line:    string(output),
			Message: fmt.Sprintf("failed to parse JSON response: %v", err),
		}
	}

	return &result, nil
}

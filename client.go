package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// Client interface for Claude Code interaction
type Client interface {
	Query(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error)
	QueryStream(ctx context.Context, prompt string, opts *Options, handler func(Message) error) error
}

// ClaudeClient implements the Client interface
type ClaudeClient struct {
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

// NewClient creates a new ClaudeClient with default components
func NewClient() *ClaudeClient {
	return &ClaudeClient{
		executor: &DefaultCommandExecutor{},
		builder:  &ArgumentBuilder{},
		parser:   &DefaultMessageParser{},
	}
}

// NewClientWithExecutor creates a new ClaudeClient with a custom executor
func NewClientWithExecutor(executor CommandExecutor) *ClaudeClient {
	return &ClaudeClient{
		executor: executor,
		builder:  &ArgumentBuilder{},
		parser:   &DefaultMessageParser{},
	}
}

// Query executes a Claude Code query and returns the result
func (c *ClaudeClient) Query(ctx context.Context, prompt string, opts *Options) (*ResultMessage, error) {
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
	output, err := c.executor.Execute(ctx, executable, args, prompt)
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

// QueryStream executes a Claude Code query with streaming output
func (c *ClaudeClient) QueryStream(ctx context.Context, prompt string, opts *Options, handler func(Message) error) error {
	if prompt == "" {
		return &ConfigError{
			Field:   "prompt",
			Message: "prompt is required",
		}
	}

	if opts == nil {
		opts = &Options{}
	}

	// Validate options
	if err := c.builder.Validate(opts); err != nil {
		return err
	}

	// Set defaults
	executable := opts.PathToClaudeCodeExecutable
	if executable == "" {
		executable = "claude"
	}

	// Build arguments
	args := append([]string{"--print", "--output-format", "stream-json", "--verbose"}, c.builder.BuildArgs(opts)...)

	// Execute command with streaming
	stream, err := c.executor.ExecuteStream(ctx, executable, args, prompt)
	if err != nil {
		if processErr, ok := err.(*ProcessError); ok {
			return processErr
		}
		return fmt.Errorf("failed to execute command: %w", err)
	}
	defer stream.Close()

	// Read streaming output
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		msg, err := c.parser.ParseMessage(line)
		if err != nil {
			return err
		}

		if err := handler(msg); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

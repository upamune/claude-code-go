package claude

import (
	"bufio"
	"context"
	"fmt"
	"io"
)

// MessageStream represents a stream of messages from Claude
type MessageStream struct {
	Messages <-chan MessageOrError
	ctx      context.Context
	cancel   context.CancelFunc
}

// MessageOrError wraps a Message or an error
type MessageOrError struct {
	Message Message
	Err     error
}

// Close cancels the stream
func (s *MessageStream) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

// QueryStream executes a Claude Code query and returns a channel of messages
func (c *ClaudeClient) QueryStream(ctx context.Context, prompt string, opts *Options) (*MessageStream, error) {
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
	args := append([]string{"--print", "--output-format", "stream-json", "--verbose"}, c.builder.BuildArgs(opts)...)

	// Create context for cancellation
	streamCtx, cancel := context.WithCancel(ctx)
	
	// Execute command with streaming
	stream, err := c.executor.ExecuteStream(streamCtx, executable, args, prompt)
	if err != nil {
		cancel()
		if processErr, ok := err.(*ProcessError); ok {
			return nil, processErr
		}
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	// Create message channel
	messages := make(chan MessageOrError)

	// Start goroutine to read messages
	go func() {
		defer close(messages)
		defer stream.Close()
		defer cancel()

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			msg, err := c.parser.ParseMessage(line)
			if err != nil {
				select {
				case messages <- MessageOrError{Err: err}:
				case <-streamCtx.Done():
					return
				}
				return
			}

			select {
			case messages <- MessageOrError{Message: msg}:
			case <-streamCtx.Done():
				return
			}
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			select {
			case messages <- MessageOrError{Err: fmt.Errorf("error reading stream: %w", err)}:
			case <-streamCtx.Done():
			}
		}
	}()

	return &MessageStream{
		Messages: messages,
		ctx:      streamCtx,
		cancel:   cancel,
	}, nil
}


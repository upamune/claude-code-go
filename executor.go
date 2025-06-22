package claude

import (
	"context"
	"io"
	"os/exec"
	"strings"
)

// CommandExecutor is an interface for executing commands
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args []string, stdin string) ([]byte, error)
	ExecuteStream(ctx context.Context, name string, args []string, stdin string) (io.ReadCloser, error)
}

// DefaultCommandExecutor implements CommandExecutor using os/exec
type DefaultCommandExecutor struct{}

// Execute runs a command and returns its output
func (e *DefaultCommandExecutor) Execute(ctx context.Context, name string, args []string, stdin string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(stdin)
	return cmd.CombinedOutput()
}

// ExecuteStream runs a command and returns a stream of its output
func (e *DefaultCommandExecutor) ExecuteStream(ctx context.Context, name string, args []string, stdin string) (io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(stdin)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &streamReader{
		reader: stdout,
		cmd:    cmd,
	}, nil
}

// streamReader wraps stdout pipe and ensures command cleanup
type streamReader struct {
	reader io.ReadCloser
	cmd    *exec.Cmd
}

func (s *streamReader) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

func (s *streamReader) Close() error {
	s.reader.Close()
	return s.cmd.Wait()
}

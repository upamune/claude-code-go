package claude

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
)

// Compile-time check that implementations satisfy the interface
var _ CommandExecutor = (*DefaultCommandExecutor)(nil)

// CommandExecutor is an interface for executing commands
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args []string, stdin string, workingDir string) ([]byte, error)
	ExecuteStream(ctx context.Context, name string, args []string, stdin string, workingDir string) (io.ReadCloser, error)
}

// DefaultCommandExecutor implements CommandExecutor using os/exec
type DefaultCommandExecutor struct{}

// Execute runs a command and returns its output
func (e *DefaultCommandExecutor) Execute(ctx context.Context, name string, args []string, stdin string, workingDir string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(stdin)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, &ProcessError{
				ExitCode: exitErr.ExitCode(),
				Message:  string(output),
			}
		}
		return nil, err
	}
	return output, nil
}

// ExecuteStream runs a command and returns a stream of its output
func (e *DefaultCommandExecutor) ExecuteStream(ctx context.Context, name string, args []string, stdin string, workingDir string) (io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(stdin)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Capture stderr for error reporting
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &streamReader{
		reader:    stdout,
		cmd:       cmd,
		stderrBuf: &stderrBuf,
	}, nil
}

// streamReader wraps stdout pipe and ensures command cleanup
type streamReader struct {
	reader    io.ReadCloser
	cmd       *exec.Cmd
	stderrBuf *bytes.Buffer
}

func (s *streamReader) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

func (s *streamReader) Close() error {
	s.reader.Close()
	err := s.cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ProcessError{
				ExitCode: exitErr.ExitCode(),
				Message:  s.stderrBuf.String(),
			}
		}
		return err
	}
	return nil
}

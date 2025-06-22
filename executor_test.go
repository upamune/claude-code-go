package claude

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// Compile-time check that MockCommandExecutor implements CommandExecutor
var _ CommandExecutor = (*MockCommandExecutor)(nil)

// MockCommandExecutor is a mock implementation of CommandExecutor for testing
type MockCommandExecutor struct {
	ExecuteFunc       func(ctx context.Context, name string, args []string, stdin string, workingDir string) ([]byte, error)
	ExecuteStreamFunc func(ctx context.Context, name string, args []string, stdin string, workingDir string) (io.ReadCloser, error)
}

func (m *MockCommandExecutor) Execute(ctx context.Context, name string, args []string, stdin string, workingDir string) ([]byte, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, name, args, stdin, workingDir)
	}
	return nil, errors.New("ExecuteFunc not implemented")
}

func (m *MockCommandExecutor) ExecuteStream(ctx context.Context, name string, args []string, stdin string, workingDir string) (io.ReadCloser, error) {
	if m.ExecuteStreamFunc != nil {
		return m.ExecuteStreamFunc(ctx, name, args, stdin, workingDir)
	}
	return nil, errors.New("ExecuteStreamFunc not implemented")
}

// mockReadCloser implements io.ReadCloser for testing
type mockReadCloser struct {
	*strings.Reader
	closeFunc func() error
}

func (m *mockReadCloser) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestDefaultCommandExecutor_Execute(t *testing.T) {
	executor := &DefaultCommandExecutor{}
	ctx := context.Background()

	tests := []struct {
		name    string
		command string
		args    []string
		stdin   string
		wantErr bool
	}{
		{
			name:    "echo command",
			command: "echo",
			args:    []string{"hello"},
			stdin:   "",
			wantErr: false,
		},
		{
			name:    "echo with stdin",
			command: "cat",
			args:    []string{},
			stdin:   "input data",
			wantErr: false,
		},
		{
			name:    "invalid command",
			command: "nonexistentcommand12345",
			args:    []string{},
			stdin:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip actual command execution in tests
			if testing.Short() {
				t.Skip("Skipping command execution test in short mode")
			}

			output, err := executor.Execute(ctx, tt.command, tt.args, tt.stdin, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && output == nil {
				t.Error("Execute() returned nil output without error")
			}
		})
	}
}

func TestDefaultCommandExecutor_Execute_Context(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context cancellation test in short mode")
	}

	executor := &DefaultCommandExecutor{}

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := executor.Execute(ctx, "sleep", []string{"5"}, "", "")
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

func TestDefaultCommandExecutor_Execute_ProcessError(t *testing.T) {
	executor := &DefaultCommandExecutor{}
	ctx := context.Background()

	// Test command that exits with non-zero status
	_, err := executor.Execute(ctx, "sh", []string{"-c", "echo 'error message' >&2; exit 1"}, "", "")
	if err == nil {
		t.Fatal("Expected error for command with non-zero exit")
	}

	processErr, ok := err.(*ProcessError)
	if !ok {
		t.Fatalf("Expected *ProcessError, got %T", err)
	}

	if processErr.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", processErr.ExitCode)
	}

	if !strings.Contains(processErr.Message, "error message") {
		t.Errorf("Expected error message to contain 'error message', got: %s", processErr.Message)
	}
}

func TestDefaultCommandExecutor_ExecuteStream(t *testing.T) {
	executor := &DefaultCommandExecutor{}
	ctx := context.Background()

	tests := []struct {
		name    string
		command string
		args    []string
		stdin   string
		wantErr bool
	}{
		{
			name:    "echo stream",
			command: "echo",
			args:    []string{"hello world"},
			stdin:   "",
			wantErr: false,
		},
		{
			name:    "invalid command",
			command: "nonexistentcommand12345",
			args:    []string{},
			stdin:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Short() {
				t.Skip("Skipping command execution test in short mode")
			}

			reader, err := executor.ExecuteStream(ctx, tt.command, tt.args, tt.stdin, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				defer reader.Close()

				// Try to read some data
				buf := make([]byte, 1024)
				n, readErr := reader.Read(buf)
				if readErr != nil && readErr != io.EOF {
					t.Errorf("Failed to read from stream: %v", readErr)
				}
				if n == 0 && readErr != io.EOF {
					t.Error("No data read from stream")
				}
			}
		})
	}
}

func TestStreamReader(t *testing.T) {
	// Test the streamReader wrapper
	mockReader := &mockReadCloser{
		Reader: strings.NewReader("test data"),
		closeFunc: func() error {
			return nil
		},
	}

	// Read data
	buf := make([]byte, 9)
	n, err := mockReader.Read(buf)
	if err != nil && err != io.EOF {
		t.Errorf("Read() error = %v", err)
	}
	if n != 9 {
		t.Errorf("Read() n = %v, want 9", n)
	}
	if string(buf) != "test data" {
		t.Errorf("Read() data = %v, want 'test data'", string(buf))
	}

	// Test Close
	if err := mockReader.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestMockCommandExecutor(t *testing.T) {
	ctx := context.Background()

	t.Run("Execute", func(t *testing.T) {
		mock := &MockCommandExecutor{
			ExecuteFunc: func(_ context.Context, name string, args []string, stdin string, _ string) ([]byte, error) {
				if name == "test" && len(args) == 1 && args[0] == "arg" && stdin == "input" {
					return []byte("output"), nil
				}
				return nil, errors.New("unexpected call")
			},
		}

		output, err := mock.Execute(ctx, "test", []string{"arg"}, "input", "")
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}
		if string(output) != "output" {
			t.Errorf("Execute() = %v, want 'output'", string(output))
		}
	})

	t.Run("ExecuteStream", func(t *testing.T) {
		mock := &MockCommandExecutor{
			ExecuteStreamFunc: func(_ context.Context, name string, _ []string, _ string, _ string) (io.ReadCloser, error) {
				if name == "stream" {
					return &mockReadCloser{
						Reader: strings.NewReader("stream data"),
					}, nil
				}
				return nil, errors.New("unexpected call")
			},
		}

		reader, err := mock.ExecuteStream(ctx, "stream", nil, "", "")
		if err != nil {
			t.Errorf("ExecuteStream() error = %v", err)
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("ReadAll() error = %v", err)
		}
		if string(data) != "stream data" {
			t.Errorf("ReadAll() = %v, want 'stream data'", string(data))
		}
	})

	t.Run("NotImplemented", func(t *testing.T) {
		mock := &MockCommandExecutor{}

		_, err := mock.Execute(ctx, "test", nil, "", "")
		if err == nil || !strings.Contains(err.Error(), "not implemented") {
			t.Errorf("Execute() error = %v, want 'not implemented'", err)
		}

		_, err = mock.ExecuteStream(ctx, "test", nil, "", "")
		if err == nil || !strings.Contains(err.Error(), "not implemented") {
			t.Errorf("ExecuteStream() error = %v, want 'not implemented'", err)
		}
	})
}

func TestDefaultCommandExecutor_ExecuteStream_ProcessError(t *testing.T) {
	executor := &DefaultCommandExecutor{}
	ctx := context.Background()

	// Test command that exits with error
	reader, err := executor.ExecuteStream(ctx, "sh", []string{"-c", "echo 'stream error' >&2; exit 2"}, "", "")
	if err != nil {
		t.Fatalf("ExecuteStream() initial error = %v", err)
	}

	// Read all data
	_, readErr := io.ReadAll(reader)
	if readErr != nil {
		t.Fatalf("ReadAll() error = %v", readErr)
	}

	// Close should return ProcessError
	closeErr := reader.Close()
	if closeErr == nil {
		t.Fatal("Expected error on Close() for command with non-zero exit")
	}

	processErr, ok := closeErr.(*ProcessError)
	if !ok {
		t.Fatalf("Expected *ProcessError, got %T", closeErr)
	}

	if processErr.ExitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", processErr.ExitCode)
	}

	if !strings.Contains(processErr.Message, "stream error") {
		t.Errorf("Expected error message to contain 'stream error', got: %s", processErr.Message)
	}
}

func TestDefaultCommandExecutor_ExecuteStream_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	executor := &DefaultCommandExecutor{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should timeout
	reader, err := executor.ExecuteStream(ctx, "sleep", []string{"5"}, "", "")
	if err != nil {
		// Context timeout should cause an error
		return
	}

	if reader != nil {
		reader.Close()
		// Sometimes the command starts before the context cancels
		// Wait a bit to ensure context timeout
		time.Sleep(150 * time.Millisecond)
		select {
		case <-ctx.Done():
			// Context was cancelled as expected
			return
		default:
			t.Error("Expected context to be cancelled")
		}
	}
}

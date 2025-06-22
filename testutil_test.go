package claude

import "testing"

// skipIfClaudeNotAvailable skips the test if Claude CLI is not available
func skipIfClaudeNotAvailable(t *testing.T) {
	t.Helper()
	if !IsClaudeAvailable() {
		t.Skip("Claude CLI not found in PATH, skipping test")
	}
}

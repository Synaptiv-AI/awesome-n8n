package condense

import (
	"os/exec"
	"testing"
)

func TestClaudeRunner_Integration(t *testing.T) {
	// Skip if claude CLI not available
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not found, skipping integration test")
	}

	result, err := ClaudeRunner("Say only the word 'hello' with no other text.", "Respond with exactly one word.")
	if err != nil {
		t.Fatalf("claude runner error: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result from claude")
	}
	t.Logf("claude response: %q", result)
}

package condense

import (
	"bytes"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"os/exec"
	"strings"
	"time"
)

const maxRetries = 3

// ClaudeRunner shells out to `claude -p` for LLM condensation.
// content is piped via stdin, prompt is passed as the -p argument.
// A system prompt frames the condensation role.
func ClaudeRunner(content, prompt string) (string, error) {
	var lastErr error

	for attempt := range maxRetries {
		result, err := runClaude(content, prompt)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// Exponential backoff with jitter
		base := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		jitter := time.Duration(rand.IntN(1000)) * time.Millisecond
		time.Sleep(base + jitter)
	}

	return "", fmt.Errorf("claude failed after %d attempts: %w", maxRetries, lastErr)
}

const systemPrompt = `You are a technical documentation condenser. You receive n8n documentation sections and condense them according to the user's instructions. Output only the condensed markdown. No commentary, no explanations, no preamble.`

func runClaude(content, prompt string) (string, error) {
	cmd := exec.Command("claude",
		"-p", prompt,
		"--system-prompt", systemPrompt,
		"--model", "haiku",
		"--output-format", "text",
		"--no-session-persistence",
	)
	cmd.Stdin = strings.NewReader(content)

	// Remove CLAUDECODE env var to allow running from within a Claude Code session
	cmd.Env = filterEnv(os.Environ(), "CLAUDECODE")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude -p failed: %w\nstderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// filterEnv returns env with the named variable removed.
func filterEnv(env []string, name string) []string {
	prefix := name + "="
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

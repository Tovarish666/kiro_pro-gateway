// Package kirocli wraps the kiro-cli binary as a subprocess.
// This is the current (v1) backend — v2 will talk to the Kiro API directly.
package kirocli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\[m`)

// Runner calls kiro-cli as a subprocess.
type Runner struct {
	BinaryPath string
}

// Run sends prompt to kiro-cli and returns the cleaned text response.
func (r *Runner) Run(apiKey, prompt string) (string, error) {
	cmd := exec.Command(r.BinaryPath, "chat", "--no-interactive")
	cmd.Env = append(os.Environ(),
		"KIRO_API_KEY="+apiKey,
		"NO_COLOR=1",
	)
	cmd.Stdin = strings.NewReader(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("kiro-cli error: %v\nstderr: %s", err, stderr.String())
	}

	return cleanOutput(stdout.String()), nil
}

func cleanOutput(raw string) string {
	out := ansiRe.ReplaceAllString(raw, "")
	lines := strings.Split(strings.TrimRight(out, "\r\n"), "\n")
	var kept []string
	for _, line := range lines {
		if strings.Contains(line, "(Credits:") {
			continue
		}
		line = strings.TrimPrefix(line, "> ")
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

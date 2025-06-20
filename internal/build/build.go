package build

import (
	"os/exec"
	"strings"
	"time"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = time.Now().Format(time.RFC3339) // Default to now, override if needed
)

func init() {
	// Try to use Git if Commit or Date still set to defaults (optional)
	if Commit == "unknown" {
		if gitCommit, err := runGitCommand("rev-parse", "--short", "HEAD"); err == nil {
			Commit = gitCommit
		}
	}
	if Version == "dev" {
		if gitTag, err := runGitCommand("describe", "--tags", "--always"); err == nil {
			Version = gitTag
		}
	}
}

// Info returns build metadata.
func Info() map[string]any {
	return map[string]any{
		"version": Version,
		"commit":  Commit,
		"date":    Date,
	}
}

// runGitCommand executes a git command and trims the output.
func runGitCommand(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

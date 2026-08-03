package main

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

// IsGitRepo checks if the current directory is a git repository.
func IsGitRepo() bool {
	cmd := exec.Command("git", "--no-pager", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

// HasStagedChanges checks if there are any staged changes.
func HasStagedChanges() bool {
	cmd := exec.Command("git", "--no-pager", "diff", "--staged", "--quiet")
	err := cmd.Run()
	// git diff --quiet returns exit status 1 if there were differences
	return err != nil
}

// GetCurrentBranch returns the current git branch name.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "--no-pager", "branch", "--show-current")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// GetFilteredStagedDiff returns the staged diff, filtering out noisy files.
func GetFilteredStagedDiff() (string, error) {
	excludes := []string{
		"go.sum",
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		"*.min.js",
		"*.min.css",
	}

	args := []string{"--no-pager", "diff", "--staged", "--"}
	args = append(args, ".")
	for _, exclude := range excludes {
		args = append(args, ":(exclude)"+exclude)
	}

	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

// GetRecentCommitMessages returns the last count commit messages formatted as string or empty string if none/error.
func GetRecentCommitMessages(count int) string {
	if count <= 0 {
		return ""
	}
	cmd := exec.Command("git", "--no-pager", "log", "-n", strconv.Itoa(count), "--format=%s", "--no-merges")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

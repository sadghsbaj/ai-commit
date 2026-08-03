package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed prompt.md
var embeddedPrompt string

func main() {
	// Pre-flight checks
	if !IsGitRepo() {
		PrintError("Not inside a valid Git repository.", nil)
		os.Exit(1)
	}

	testPrompt := false
	for _, arg := range os.Args[1:] {
		if arg == "--test-prompt" || arg == "-t" || arg == "--dry-run" {
			testPrompt = true
			break
		}
	}

	if !testPrompt && !HasStagedChanges() {
		PrintError("No staged changes found. Please run 'git add' to stage your changes first.", nil)
		os.Exit(1)
	}

	err := LoadConfig()
	if err != nil {
		PrintError("Could not load config", err)
	}

	systemPrompt := strings.TrimSpace(embeddedPrompt)
	if systemPrompt == "" {
		systemPrompt = "You are an expert software engineer. Analyze the provided git diff and the current branch name. Write a professional, technically accurate, and concise git commit message in English. Use the Conventional Commits format (e.g., feat:, fix:, chore:). Only return the commit message, no markdown formatting, no explanations."
	}

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		promptPath := filepath.Join(exeDir, "prompt.md")
		if systemPromptBytes, err := os.ReadFile(promptPath); err == nil && len(strings.TrimSpace(string(systemPromptBytes))) > 0 {
			systemPrompt = string(systemPromptBytes)
		}
	}

	branch, _ := GetCurrentBranch()
	diff, err := GetFilteredStagedDiff()
	if err != nil && !testPrompt {
		PrintError("Failed to get staged diff", err)
		os.Exit(1)
	}

	if testPrompt {
		if strings.TrimSpace(diff) == "" {
			diff = "<no staged changes / example diff>"
		}
		fmt.Println(BuildPromptText(systemPrompt, diff, branch, ""))
		return
	}

	if diff == "" {
		// Just to be completely sure diff filtering didn't hide everything
		PrintError("Filtered diff is empty. All staged changes are excluded.", nil)
		os.Exit(1)
	}

	userComment := ""

	for {
		spinner := StartSpinner("⏳ Thinking...")
		suggestion, err := GenerateCommitMessage(context.Background(), systemPrompt, diff, branch, userComment)
		spinner.Stop()

		if err != nil {
			PrintError("Failed to generate commit message", err)
			os.Exit(1)
		}

		SendDesktopNotification("AI Commit", "Vorschlag ist da!")

		action, nextComment := PromptUser(suggestion)

		if action == 'a' {
			// Commit
			cmd := exec.Command("git", "commit", "-m", suggestion)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				PrintError("Failed to execute git commit", err)
				os.Exit(1)
			}
			PrintSuccess("Commit created successfully!")
			break
		} else if action == 'r' {
			fmt.Println("Aborted by user.")
			break
		} else if action == 'c' {
			userComment = nextComment
			// loop continues
		}
	}
}

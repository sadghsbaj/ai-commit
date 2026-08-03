package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

// GeminiRequest represents the payload to the Gemini API
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// LoadConfig loads the environment variables
func LoadConfig() error {
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		envPath := filepath.Join(exeDir, ".env")
		err := godotenv.Load(envPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error loading .env file: %w", err)
		}
	}
	// It's strictly okay if .env doesn't exist, it might be in the system env
	return nil
}

// BuildPromptText constructs the full prompt payload string sent to the LLM.
func BuildPromptText(systemPrompt, diff, branch, userComment, recentCommits string) string {
	promptText := systemPrompt + "\n\nBranch: " + branch
	if recentCommits != "" {
		promptText += "\n\nRecent Commit History:\n" + recentCommits
	}
	promptText += "\n\nDiff:\n" + diff
	if userComment != "" {
		promptText += "\n\nUser Feedback for refinement: " + userComment
	}
	return promptText
}

// GenerateCommitMessage calls the Gemini API to get a commit message.
func GenerateCommitMessage(ctx context.Context, systemPrompt, diff, branch, userComment, recentCommits string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	modelName := os.Getenv("AI_COMMIT_MODEL")

	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not set. Please set it in .env or your environment")
	}
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, apiKey)

	promptText := BuildPromptText(systemPrompt, diff, branch, userComment, recentCommits)

	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{Text: promptText},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	return executeWithRetries(ctx, url, jsonData)
}

func executeWithRetries(ctx context.Context, url string, jsonData []byte) (string, error) {
	maxRetries := 5
	backoff := 1 * time.Second

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)

		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					return "", fmt.Errorf("failed to read response: %w", err)
				}
				var geminiResp GeminiResponse
				if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
					return "", fmt.Errorf("failed to parse response: %w", err)
				}
				if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
					return geminiResp.Candidates[0].Content.Parts[0].Text, nil
				}
				return "", fmt.Errorf("unexpected response format")
			} else if resp.StatusCode == 429 || resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("API error: status code %d", resp.StatusCode)
			} else {
				// Don't retry on 400s (except 429)
				bodyBytes, _ := io.ReadAll(resp.Body)
				return "", fmt.Errorf("API error: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
			}
		} else {
			lastErr = err
		}

		if attempt < maxRetries {
			select {
			case <-time.After(backoff):
				backoff *= 2
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}

	return "", fmt.Errorf("failed after %d retries. Last error: %w", maxRetries, lastErr)
}

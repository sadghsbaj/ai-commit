package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExecuteWithRetries_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GeminiResponse{}
		resp.Candidates = make([]struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		}, 1)
		resp.Candidates[0].Content.Parts = make([]struct {
			Text string `json:"text"`
		}, 1)
		resp.Candidates[0].Content.Parts[0].Text = "feat: add awesome feature"

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := executeWithRetries(ctx, server.URL, []byte(`{}`))
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if result != "feat: add awesome feature" {
		t.Errorf("Expected 'feat: add awesome feature', got '%s'", result)
	}
}

func TestExecuteWithRetries_RetryAndFail(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Short timeout to avoid long test waiting for all 5 retries (1,2,4,8,16) -> 31s
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := executeWithRetries(ctx, server.URL, []byte(`{}`))
	if err == nil {
		t.Fatalf("Expected error, got none")
	}

	if attempts < 2 {
		t.Errorf("Expected multiple attempts, got %d", attempts)
	}
}

func TestExecuteWithRetries_RetryAndSuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		resp := GeminiResponse{}
		resp.Candidates = make([]struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		}, 1)
		resp.Candidates[0].Content.Parts = make([]struct {
			Text string `json:"text"`
		}, 1)
		resp.Candidates[0].Content.Parts[0].Text = "feat: accepted after retry"

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := executeWithRetries(ctx, server.URL, []byte(`{}`))
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if result != "feat: accepted after retry" {
		t.Errorf("Expected 'feat: accepted after retry', got '%s'", result)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

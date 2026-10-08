package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var baseURL = "http://localhost:8080/api/v1"

func init() {
	if url := os.Getenv("SNIPO_E2E_API_URL"); url != "" {
		baseURL = url
	}
}

// waitForServer polls the health endpoint until it responds or times out
func waitForServer(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("Server did not become ready at %s within 30 seconds", baseURL)
}

func TestE2E_HealthCheck(t *testing.T) {
	waitForServer(t)

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("status\": \"ok")) {
		t.Errorf("Expected healthy status, got %s", string(body))
	}
}

func TestE2E_SnippetFlow(t *testing.T) {
	waitForServer(t)
	client := &http.Client{}

	// Create snippet
	createPayload := map[string]string{
		"title":       "E2E Test Snippet",
		"content":     "echo 'Hello World'",
		"language":    "bash",
		"description": "Created via E2E tests",
	}
	bodyBytes, _ := json.Marshal(createPayload)

	req, _ := http.NewRequest("POST", baseURL+"/snippets", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to create snippet: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 201 or 401 (if auth required), got %d", resp.StatusCode)
	}
}

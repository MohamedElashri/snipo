package main_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTUI_E2E(t *testing.T) {
	// Build the binary first to avoid 'go run' leaving orphaned child processes
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "snippy")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/snippy")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build TUI binary: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath)
	cmd.Env = append(os.Environ(), "SNIPO_API_URL=http://localhost:8080")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start TUI process: %v", err)
	}

	err = cmd.Wait()

	outStr := stdout.String()
	errStr := stderr.String()

	isTimeout := ctx.Err() == context.DeadlineExceeded || (err != nil && err.Error() == "signal: killed")
	
	hasEscapeSequences := strings.Contains(outStr, "\x1b") || strings.Contains(errStr, "\x1b")
	hasConnError := strings.Contains(strings.ToLower(outStr+errStr), "connection refused") || 
		strings.Contains(strings.ToLower(outStr+errStr), "dial tcp") ||
		strings.Contains(strings.ToLower(outStr+errStr), "error")

	if !hasEscapeSequences && !hasConnError {
		t.Errorf("Expected UI initialization bytes or connection error, but got:\nError: %v\nTimeout: %v\nStdout: %q\nStderr: %q", 
			err, isTimeout, outStr, errStr)
	}
}

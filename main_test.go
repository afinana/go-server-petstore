package main

import (
	"os/exec"
	"testing"
)

// Ensure the module builds (compiles) — does not run the server.
func TestBuildProject(t *testing.T) {
	cmd := exec.Command("go", "build", "./...")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}
}

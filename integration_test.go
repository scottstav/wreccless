package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Build ccl binary
	binDir := t.TempDir()
	cclBin := filepath.Join(binDir, "ccl")
	buildCmd := exec.Command("go", "build", "-o", cclBin, "./cmd/ccl")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %s\n%s", err, out)
	}

	stateDir := t.TempDir()
	configDir := t.TempDir()

	configPath := filepath.Join(configDir, "config.toml")
	os.WriteFile(configPath, []byte("[claude]\nskip_permissions = true\n"), 0644)

	env := append(os.Environ(),
		"CCL_STATE_DIR="+stateDir,
		"CCL_CONFIG="+configPath,
	)

	run := func(args ...string) (string, error) {
		cmd := exec.Command(cclBin, args...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}

	// 1. Create a pending worker
	out, err := run("new", "--dir", t.TempDir(), "--task", "integration test", "--pending")
	if err != nil {
		t.Fatalf("new --pending: %s (%v)", out, err)
	}
	workerID := out
	if workerID == "" {
		t.Fatal("expected worker ID")
	}
	t.Logf("Created pending worker: %s", workerID)

	// 2. List should show it as pending
	out, err = run("list", "--json")
	if err != nil {
		t.Fatalf("list: %s (%v)", out, err)
	}
	var workers []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &workers); err != nil {
		t.Fatalf("json parse: %v (%s)", err, out)
	}
	if len(workers) != 1 || workers[0]["status"] != "pending" {
		t.Errorf("expected 1 pending worker: %s", out)
	}

	// 3. Status should show details
	out, err = run("status", workerID, "--json")
	if err != nil {
		t.Fatalf("status: %s (%v)", out, err)
	}
	var statusResult map[string]interface{}
	if err := json.Unmarshal([]byte(out), &statusResult); err != nil {
		t.Fatalf("status json: %v", err)
	}
	if statusResult["status"] != "pending" {
		t.Errorf("expected pending status: %v", statusResult["status"])
	}

	// 4. Deny the worker
	out, err = run("deny", workerID)
	if err != nil {
		t.Fatalf("deny: %s (%v)", out, err)
	}
	t.Logf("Denied: %s", out)

	// 5. List should be empty
	out, _ = run("list")
	if !strings.Contains(out, "No workers") {
		t.Errorf("expected no workers after deny: %s", out)
	}

	// 6. Create another pending, then approve it
	out, err = run("new", "--dir", t.TempDir(), "--task", "approve test", "--pending")
	if err != nil {
		t.Fatalf("new --pending (2): %s (%v)", out, err)
	}
	workerID2 := out

	// Approve will try to spawn ccl run, which will fail since there's no real claude
	// but the state should transition to working
	out, err = run("approve", workerID2)
	if err != nil {
		// The approve might fail because ccl run spawns and immediately fails
		// That's OK — we just want to verify the state transition
		t.Logf("approve output (may have spawn error): %s %v", out, err)
	}

	// Check state is working (approve changes it before spawning)
	out, _ = run("status", workerID2, "--json")
	if out != "" {
		var s2 map[string]interface{}
		json.Unmarshal([]byte(out), &s2)
		if s2["status"] == "pending" {
			t.Error("worker should not be pending after approve")
		}
	}

	// 7. Clean up
	run("clean", "--all")
	out, _ = run("list")
	if !strings.Contains(out, "No workers") {
		t.Errorf("expected no workers after clean --all: %s", out)
	}

	t.Log("Integration lifecycle test passed")
}

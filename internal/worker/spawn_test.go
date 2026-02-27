package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSpawnDetached(t *testing.T) {
	binDir := t.TempDir()
	mockBin := filepath.Join(binDir, "ccl")
	marker := filepath.Join(t.TempDir(), "spawned")
	os.WriteFile(mockBin, []byte(fmt.Sprintf("#!/bin/sh\ntouch %s\n", marker)), 0755)

	err := SpawnRun("1234", mockBin, "/tmp/config", "/tmp/state")
	if err != nil {
		t.Fatalf("SpawnRun: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		t.Error("detached process did not run — marker file missing")
	}
}

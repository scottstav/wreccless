package state

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStateDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func TestWriteAndRead(t *testing.T) {
	dir := tempStateDir(t)
	w := &Worker{
		ID:        "123",
		Status:    StatusWorking,
		Directory: "/tmp/foo",
		Task:      "fix bug",
		PID:       42,
		SessionID: "abc-def",
	}
	if err := Write(dir, w); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got, err := Read(dir, "123")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.ID != "123" || got.Status != StatusWorking || got.Task != "fix bug" {
		t.Errorf("unexpected worker: %+v", got)
	}
	if got.PID != 42 || got.SessionID != "abc-def" {
		t.Errorf("unexpected PID/SessionID: %+v", got)
	}
}

func TestReadNotFound(t *testing.T) {
	dir := tempStateDir(t)
	_, err := Read(dir, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent worker")
	}
}

func TestList(t *testing.T) {
	dir := tempStateDir(t)
	for _, id := range []string{"100", "200", "300"} {
		w := &Worker{ID: id, Status: StatusDone, Directory: "/tmp", Task: "task " + id}
		if err := Write(dir, w); err != nil {
			t.Fatalf("Write %s: %v", id, err)
		}
	}
	workers, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(workers) != 3 {
		t.Errorf("expected 3 workers, got %d", len(workers))
	}
}

func TestDelete(t *testing.T) {
	dir := tempStateDir(t)
	w := &Worker{ID: "456", Status: StatusPending, Directory: "/tmp", Task: "test"}
	Write(dir, w)
	if err := Delete(dir, "456"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := Read(dir, "456")
	if err == nil {
		t.Fatal("expected error after delete")
	}
	// Verify log file also deleted
	logPath := filepath.Join(dir, "456.log")
	os.WriteFile(logPath, []byte("log"), 0644)
	Write(dir, &Worker{ID: "456", Status: StatusDone, Directory: "/tmp", Task: "t"})
	os.WriteFile(logPath, []byte("log"), 0644)
	Delete(dir, "456")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("log file should be deleted too")
	}
}

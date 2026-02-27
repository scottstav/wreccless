package hooks

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRenderTemplate(t *testing.T) {
	vars := Vars{
		ID:        "123",
		Task:      "fix the bug",
		Dir:       "/home/user/project",
		Status:    "done",
		SessionID: "abc-def",
	}
	result, err := render("Worker {{.ID}} is {{.Status}}: {{.Task}}", vars)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if result != "Worker 123 is done: fix the bug" {
		t.Errorf("unexpected: %q", result)
	}
}

func TestFireCreatesMarkerFile(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "fired")
	cmds := []string{"touch " + marker}
	vars := Vars{ID: "1", Task: "t", Dir: "/tmp", Status: "done"}
	Fire(cmds, vars)
	time.Sleep(200 * time.Millisecond)
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		t.Error("hook did not fire — marker file missing")
	}
}

func TestFireWithTemplateVars(t *testing.T) {
	dir := t.TempDir()
	outfile := filepath.Join(dir, "out")
	cmds := []string{"echo '{{.ID}}:{{.Status}}' > " + outfile}
	vars := Vars{ID: "42", Task: "t", Dir: "/tmp", Status: "done"}
	Fire(cmds, vars)
	time.Sleep(200 * time.Millisecond)
	data, err := os.ReadFile(outfile)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	got := string(data)
	if got != "42:done\n" {
		t.Errorf("unexpected: %q", got)
	}
}

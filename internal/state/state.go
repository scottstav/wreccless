package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusWorking Status = "working"
	StatusDone    Status = "done"
	StatusError   Status = "error"
)

type Worker struct {
	ID         string     `json:"id"`
	Status     Status     `json:"status"`
	Directory  string     `json:"directory"`
	Task       string     `json:"task"`
	Image      string     `json:"image,omitempty"`
	PID        int        `json:"pid,omitempty"`
	SessionID  string     `json:"session_id,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

func statePath(dir, id string) string {
	return filepath.Join(dir, id+".json")
}

func Write(dir string, w *Worker) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}
	tmp := statePath(dir, w.ID) + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, statePath(dir, w.ID))
}

func Read(dir, id string) (*Worker, error) {
	data, err := os.ReadFile(statePath(dir, id))
	if err != nil {
		return nil, fmt.Errorf("worker %s: %w", id, err)
	}
	var w Worker
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("worker %s: %w", id, err)
	}
	return &w, nil
}

func List(dir string) ([]*Worker, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var workers []*Worker
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		w, err := Read(dir, id)
		if err != nil {
			continue
		}
		workers = append(workers, w)
	}
	sort.Slice(workers, func(i, j int) bool {
		return workers[i].ID < workers[j].ID
	})
	return workers, nil
}

func Delete(dir, id string) error {
	os.Remove(filepath.Join(dir, id+".log"))
	return os.Remove(statePath(dir, id))
}

package tui

import (
	"encoding/json"
	"os"
)

const maxHistory = 50

func loadDirHistory(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var dirs []string
	if err := json.Unmarshal(data, &dirs); err != nil {
		return nil
	}
	return dirs
}

func saveDirHistory(path string, dirs []string) error {
	if len(dirs) > maxHistory {
		dirs = dirs[:maxHistory]
	}
	data, err := json.Marshal(dirs)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

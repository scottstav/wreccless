package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if !cfg.Claude.SkipPermissions {
		t.Error("skip_permissions should default to true")
	}
	if cfg.Claude.SystemPrompt == "" {
		t.Error("system_prompt should have a default")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`
[claude]
skip_permissions = false
system_prompt = "custom prompt"
extra_flags = ["--model", "opus"]

[hooks]
on_done = ["echo done"]
`), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Claude.SkipPermissions {
		t.Error("skip_permissions should be false")
	}
	if cfg.Claude.SystemPrompt != "custom prompt" {
		t.Errorf("system_prompt: %q", cfg.Claude.SystemPrompt)
	}
	if len(cfg.Claude.ExtraFlags) != 2 {
		t.Errorf("extra_flags: %v", cfg.Claude.ExtraFlags)
	}
	if len(cfg.Hooks.OnDone) != 1 || cfg.Hooks.OnDone[0] != "echo done" {
		t.Errorf("on_done: %v", cfg.Hooks.OnDone)
	}
}

func TestLoadMissing(t *testing.T) {
	cfg, err := Load("/nonexistent/config.toml")
	if err != nil {
		t.Fatalf("Load should not error on missing file: %v", err)
	}
	if !cfg.Claude.SkipPermissions {
		t.Error("missing file should return defaults")
	}
}

package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type ClaudeConfig struct {
	SkipPermissions bool     `toml:"skip_permissions"`
	SystemPrompt    string   `toml:"system_prompt"`
	ExtraFlags      []string `toml:"extra_flags"`
}

type HooksConfig struct {
	OnStart   []string `toml:"on_start"`
	OnDone    []string `toml:"on_done"`
	OnPending []string `toml:"on_pending"`
	OnError   []string `toml:"on_error"`
	OnKill    []string `toml:"on_kill"`
}

type Config struct {
	Claude ClaudeConfig `toml:"claude"`
	Hooks  HooksConfig  `toml:"hooks"`
}

const defaultSystemPrompt = `You are the user's trusted programmer. Do not ask questions. Complete the entire task before stopping. If you encounter issues, debug and fix them. When finished, end with a 1-2 sentence summary.`

func Defaults() *Config {
	return &Config{
		Claude: ClaudeConfig{
			SkipPermissions: true,
			SystemPrompt:    defaultSystemPrompt,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds per-repo settings.
type Config struct {
	BaseBranch   string `json:"base_branch"`    // default branch to create worktrees from (e.g. "main", "develop")
	WorktreeBase string `json:"worktree_base"`  // override worktree directory
}

// configFileName is the config file placed in the repo root.
const configFileName = ".flow.json"

// Load reads config from .flow.json in the given repo root.
// Returns defaults if file doesn't exist.
func Load(repoRoot string) Config {
	cfg := Config{
		BaseBranch: "HEAD",
	}

	data, err := os.ReadFile(filepath.Join(repoRoot, configFileName))
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)

	if cfg.BaseBranch == "" {
		cfg.BaseBranch = "HEAD"
	}
	return cfg
}

// Save writes config to .flow.json in the given repo root.
func Save(repoRoot string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(repoRoot, configFileName), data, 0o644)
}

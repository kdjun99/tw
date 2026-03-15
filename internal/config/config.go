package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Project struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	DefaultBranch string `json:"defaultBranch"`
	WorktreeDir   string `json:"worktreeDir,omitempty"`
}

func (p Project) ResolveWorktreeDir() string {
	if p.WorktreeDir != "" {
		return p.WorktreeDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tw", p.Name)
}

type Config struct {
	Projects []Project `json:"projects"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "tw")
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

func Load() (*Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(configPath(), data, 0o644)
}

func (c *Config) FindProject(name string) *Project {
	for i := range c.Projects {
		if c.Projects[i].Name == name {
			return &c.Projects[i]
		}
	}
	return nil
}

func (c *Config) AddProject(p Project) error {
	if c.FindProject(p.Name) != nil {
		return fmt.Errorf("project %q already exists", p.Name)
	}
	c.Projects = append(c.Projects, p)
	return c.Save()
}

func (c *Config) RemoveProject(name string) error {
	for i, p := range c.Projects {
		if p.Name == name {
			c.Projects = append(c.Projects[:i], c.Projects[i+1:]...)
			return c.Save()
		}
	}
	return fmt.Errorf("project %q not found", name)
}

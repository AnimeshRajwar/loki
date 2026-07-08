package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	RepoConfigPath = ".loki/config"
)

// SystemConfigPath returns the OS-appropriate system config location.
func SystemConfigPath() string {
	switch runtime.GOOS {
	case "windows":
		if pd := os.Getenv("PROGRAMDATA"); pd != "" {
			return filepath.Join(pd, "loki", "config")
		}
		return filepath.Join(string(os.PathSeparator), "ProgramData", "loki", "config")
	case "darwin":
		return filepath.Join(string(os.PathSeparator), "Library", "Application Support", "loki", "config")
	default:
		return filepath.Join(string(os.PathSeparator), "etc", "loki", "config")
	}
}

type Config struct {
	values map[string]string
}

func NewConfig() *Config {
	return &Config{values: make(map[string]string)}
}

func (c *Config) Load(repoRoot string) error {
	// System
	c.loadFile(SystemConfigPath())
	// User
	// Load XDG user config (if XDG_CONFIG_HOME set) then fallback global file
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		c.loadFile(filepath.Join(xdg, "loki", "config"))
	}
	// fallback
	c.loadFile(GlobalFallbackPath())
	// Repo
	if repoRoot != "" {
		repoPath := filepath.Join(repoRoot, RepoConfigPath)
		c.loadFile(repoPath)
	}
	return nil
}

func GlobalConfigPath() string {
	// If XDG_CONFIG_HOME is set, prefer $XDG_CONFIG_HOME/loki/config
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "loki", "config")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".lokiconfig")
	}
	return filepath.Join(homeDir, ".lokiconfig")
}

// GlobalFallbackPath returns the legacy fallback user config (~/.lokiconfig)
func GlobalFallbackPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".lokiconfig")
	}
	return filepath.Join(homeDir, ".lokiconfig")
}

func LocalConfigPath(repoRoot string) string {
	return filepath.Join(repoRoot, RepoConfigPath)
}

func (c *Config) LoadPath(path string) {
	c.loadFile(path)
}

func (c *Config) loadFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			c.values[key] = val
		}
	}
}

func (c *Config) Get(key string) string {
	return c.values[key]
}

func (c *Config) Set(level, repoRoot, key, value string) error {
	var path string
	if level == "system" {
		path = SystemConfigPath()
	} else if level == "global" {
		// Prefer XDG path when XDG_CONFIG_HOME is set, otherwise use fallback
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			path = filepath.Join(xdg, "loki", "config")
		} else {
			path = GlobalFallbackPath()
		}
	} else if level == "local" {
		if repoRoot == "" {
			return errors.New("repo root required for local config")
		}
		path = LocalConfigPath(repoRoot)
	} else {
		return errors.New("invalid config level")
	}
	return setConfigValue(path, key, value)
}

func setConfigValue(path, key, value string) error {

	lines := []string{}
	found := false
	if file, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, key+"=") {
				lines = append(lines, key+"="+value)
				found = true
			} else {
				lines = append(lines, line)
			}
		}
		file.Close()
	}
	if !found {
		lines = append(lines, key+"="+value)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, line := range lines {
		file.WriteString(line + "\n")
	}
	return nil
}

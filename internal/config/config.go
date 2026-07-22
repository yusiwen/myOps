package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type GiteaConfig struct {
	URL          string `yaml:"url"           env:"GITEA_URL"`
	Token        string `yaml:"token"         env:"GITEA_TOKEN"`
	ClientID     string `yaml:"client_id"     env:"GITEA_CLIENT_ID"`
	ClientSecret string `yaml:"client_secret" env:"GITEA_CLIENT_SECRET"`
}

type DroneConfig struct {
	URL   string `yaml:"url"   env:"DRONE_URL"`
	Token string `yaml:"token" env:"DRONE_TOKEN"`
}

type Config struct {
	ListenAddr    string      `yaml:"listen_addr"    env:"LISTEN_ADDR"`
	BaseURL       string      `yaml:"base_url"       env:"BASE_URL"`
	LogFile       string      `yaml:"log_file"       env:"LOG_FILE"`
	SessionKey    string      `yaml:"session_key"    env:"SESSION_KEY"`
	UseTreeSitter bool        `yaml:"use_tree_sitter" env:"USE_TREE_SITTER"`
	Gitea         GiteaConfig `yaml:"gitea"`
	Drone         DroneConfig `yaml:"drone"`
}

func Load(explicitPath string) (*Config, error) {
	cfg := &Config{
		ListenAddr: ":9090",
		LogFile:    "myops.log",
	}

	path, required := resolveConfigPath(explicitPath)
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !required && errors.Is(err, os.ErrNotExist) {
				goto done
			}
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

done:
	cfg.overrideFromEnv()
	return cfg, nil
}

func resolveConfigPath(explicit string) (path string, required bool) {
	if explicit != "" {
		return explicit, true
	}
	if v := os.Getenv("CONFIG_FILE"); v != "" {
		return v, true
	}
	for _, p := range searchPaths() {
		if _, err := os.Stat(p); err == nil {
			return p, false
		}
	}
	return "", false
}

func searchPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		"config.yaml",
		filepath.Join(home, ".config", "myops", "config.yaml"),
		"/etc/myops/config.yaml",
	}
}

func (c *Config) overrideFromEnv() {
	if v := os.Getenv("LISTEN_ADDR"); v != "" {
		c.ListenAddr = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		c.BaseURL = v
	}
	if v := os.Getenv("LOG_FILE"); v != "" {
		c.LogFile = v
	}
	if v := os.Getenv("SESSION_KEY"); v != "" {
		c.SessionKey = v
	}
	if v := os.Getenv("GITEA_URL"); v != "" {
		c.Gitea.URL = v
	}
	if v := os.Getenv("GITEA_TOKEN"); v != "" {
		c.Gitea.Token = v
	}
	if v := os.Getenv("GITEA_CLIENT_ID"); v != "" {
		c.Gitea.ClientID = v
	}
	if v := os.Getenv("GITEA_CLIENT_SECRET"); v != "" {
		c.Gitea.ClientSecret = v
	}
	if v := os.Getenv("DRONE_URL"); v != "" {
		c.Drone.URL = v
	}
	if v := os.Getenv("DRONE_TOKEN"); v != "" {
		c.Drone.Token = v
	}
	if v := os.Getenv("USE_TREE_SITTER"); v == "true" {
		c.UseTreeSitter = true
	}
}

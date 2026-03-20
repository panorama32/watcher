package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	SlackUserToken string `toml:"slack_user_token"`
	FetchInterval  string `toml:"fetch_interval"` // e.g. "3m", "30s"
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config %s: %w", path, err)
	}

	if cfg.SlackUserToken == "" {
		return nil, fmt.Errorf("slack_user_token is not set in %s", path)
	}

	return &cfg, nil
}

func Dir() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "watcher"), nil
}

// Path returns the path to config.toml.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func configPath() (string, error) {
	return Path()
}

// Exists returns true if config.toml exists.
func Exists() (bool, error) {
	path, err := Path()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// InitFile creates a config.toml with a template if it doesn't exist.
func InitFile() error {
	exists, err := Exists()
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("config file already exists")
	}

	path, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	template := `# Watcher configuration
slack_user_token = ""
# fetch_interval = "3m"
`
	return os.WriteFile(path, []byte(template), 0o644)
}

// ValidKeys returns the list of configurable keys.
func ValidKeys() []string {
	return []string{"slack_user_token", "fetch_interval"}
}

// LoadRaw loads config.toml as a map.
func LoadRaw() (map[string]interface{}, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if _, err := toml.DecodeFile(path, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// SaveRaw writes a map back to config.toml.
func SaveRaw(data map[string]interface{}) error {
	path, err := Path()
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(data)
}

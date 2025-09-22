package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// Config holds application configuration values.
type Config struct {
	URL  string `json:"db_url"`
	User string `json:"current_user_name"`
}

// Read reads the JSON configuration file located at ~/.gatorconfig.json
// and decodes it into a Config. If the file does not exist an error is returned.
func Read() (Config, error) {
	var cfg Config

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg, err
	}

	path := filepath.Join(home, ".gatorconfig.json")

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (c Config) SetUser(name string) error {
	c.User = name

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	path := filepath.Join(home, ".gatorconfig.json")

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

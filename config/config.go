package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	LastFmSessionKey string `json:"lastfm_session_key"`
	LastFmUsername   string `json:"lastfm_username"`
	LastFmApiKey     string `json:"lastfm_apikey"`
	LastFmSecret     string `json:"lastfm_secret"`

	mu sync.RWMutex `json:"-"`
}

func GetConfigPath() string {
	userConfigDir, _ := os.UserConfigDir()
	dir := filepath.Join(userConfigDir, "scrobbler")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "config.json")
}

func SaveConfig(c *Config) error {
	c.mu.RLock() // Lock for reading so we dont save halfway through a change
	defer c.mu.RUnlock()

	data, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(GetConfigPath(), data, 0644)
}

// Returns a pointer to the config
func LoadConfig() (*Config, error) {
	c := &Config{}
	data, err := os.ReadFile(GetConfigPath())

	if err != nil {
		return c, err
	}

	err = json.Unmarshal(data, c)
	return c, err
}

func (c *Config) GetSessionKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastFmSessionKey
}

func (c *Config) GetUsername() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastFmUsername
}

func (c *Config) SetSessionKey(key string) error {
	c.mu.Lock()
	c.LastFmSessionKey = key
	c.mu.Unlock()

	return SaveConfig(c)
}

func (c *Config) SetUsername(user string) error {
	c.mu.Lock()
	c.LastFmUsername = user
	c.mu.Unlock()

	return SaveConfig(c)
}

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
	DiscordAppID     string `json:"discord_app_id"`
	DoPresence       bool   `json:"discord_presence_toggle"`

	mu sync.RWMutex `json:"-"`
}

func CreateDefaultConfig(path string) (*Config, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Define your default values
	defaultConf := &Config{
		LastFmSessionKey: "",
		LastFmUsername:   "",
		LastFmApiKey:     "",
		LastFmSecret:     "",
		DiscordAppID:     "",
		DoPresence:       false,
	}

	data, _ := json.MarshalIndent(defaultConf, "", "  ")
	err := os.WriteFile(path, data, 0644)

	return defaultConf, err
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
		if os.IsNotExist(err) {
			return CreateDefaultConfig(GetConfigPath())
		}
		return nil, err
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

func (c *Config) GetApiKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastFmApiKey
}

func (c *Config) GetSecrect() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastFmSecret
}

func (c *Config) GetDiscordAppId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DiscordAppID
}

func (c *Config) GetPresenceCheck() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DoPresence
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

func (c *Config) SetApiKey(apikey string) error {
	c.mu.Lock()
	c.LastFmApiKey = apikey
	c.mu.Unlock()

	return SaveConfig(c)
}

func (c *Config) SetSecret(secret string) error {
	c.mu.Lock()
	c.LastFmSecret = secret
	c.mu.Unlock()

	return SaveConfig(c)
}

func (c *Config) SetDiscordAppID(appid string) error {
	c.mu.Lock()
	c.DiscordAppID = appid
	c.mu.Unlock()

	return SaveConfig(c)
}

func (c *Config) SetPresenceCheck(doPresence bool) error {
	c.mu.Lock()
	c.DoPresence = doPresence
	c.mu.Unlock()

	return SaveConfig(c)
}

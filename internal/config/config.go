package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
)

type Config struct {
	DeviceID     string `json:"device_id"`
	DeviceName   string `json:"device_name"`
	Username     string `json:"username"`
	DownloadDir  string `json:"download_dir"`
	TCPListenPort int   `json:"tcp_listen_port"`
}

func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return createDefault()
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func createDefault() (*Config, error) {
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if username == "" {
		username = "user"
	}

	downloadDir, _ := os.UserHomeDir()
	downloadDir = filepath.Join(downloadDir, "Downloads", "LANChat")

	cfg := &Config{
		DeviceID:      generateDeviceID(),
		DeviceName:    hostname,
		Username:      username,
		DownloadDir:   downloadDir,
		TCPListenPort: 9527,
	}

	if err := cfg.Save(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configDir, "lanchat")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, "config.json"), nil
}

func generateDeviceID() string {
	return uuid.New().String()
}

func GetPlatform() string {
	return runtime.GOOS
}

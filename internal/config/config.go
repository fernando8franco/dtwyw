package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	dtwywDir       = "dtwyw"
	configFileName = "config.json"
)

type Config struct {
	APIKeys []KeyInfo `json:"api_keys"`
}

type KeyInfo struct {
	Key    string `json:"key"`
	Token  string `json:"token"`
	Status bool   `json:"status"`
}

func Read() (Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	jsonFile, err := os.Open(configFilePath)
	if err != nil {
		return Config{}, err
	}
	defer jsonFile.Close()

	var cfg Config
	if err := json.NewDecoder(jsonFile).Decode(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func write(cfg Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	configFile, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")

	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	return nil
}

func (c *Config) GetKeyInfo() KeyInfo {
	var activeKey int

	for i, k := range c.APIKeys {
		if k.Status {
			activeKey = i
			break
		}
	}

	c.SetStatus(activeKey)
	return c.APIKeys[activeKey]
}

func (c *Config) SetStatus(activeKey int) error {
	for index := range c.APIKeys {
		if index == activeKey {
			c.APIKeys[index].Status = true
		} else {
			c.APIKeys[index].Status = false
		}
	}
	return write(*c)
}

func (c *Config) GetToken() string {
	for _, k := range c.APIKeys {
		if k.Status {
			return k.Token
		}
	}

	return ""
}

func (c *Config) SetToken(key string, token string) error {
	for i, k := range c.APIKeys {
		if k.Key == key {
			c.APIKeys[i].Token = token
			break
		}
	}

	return write(*c)
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configFilePath := filepath.Join(homeDir, dtwywDir, configFileName)
	return configFilePath, nil
}

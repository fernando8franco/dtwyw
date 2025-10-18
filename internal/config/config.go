package config

import (
	"encoding/json"
	"os"
)

const configFileName = "config.json"

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

	if err := json.NewEncoder(configFile).Encode(cfg); err != nil {
		return err
	}

	return nil
}

func (c *Config) SetToken(index int, token string) error {
	c.APIKeys[index].Token = token
	return write(*c)
}

func (c *Config) GetToken(key string) error {
	for index, key := range c.APIKeys {
		if key.Status {
			activeKey = index
			break
		}
	}
	return write(*c)
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

func (c *Config) GetActiveKey() string {
	var activeKey int
	for index, key := range c.APIKeys {
		if key.Status {
			activeKey = index
			break
		}
	}

	c.SetStatus(activeKey)
	return c.APIKeys[activeKey].Key
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configFilePath := homeDir + "/dtwyw/" + configFileName
	return configFilePath, nil
}

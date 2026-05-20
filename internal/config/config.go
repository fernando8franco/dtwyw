package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	configDir      = "pressgo"
	configFileName = ".config.json"
)

type Config struct {
	Credentials map[string]Credential `json:"credentials"`
}

type Credential struct {
	Key     string `json:"key"`
	Token   string `json:"token"`
	Credits int    `json:"credits"`
	Status  bool   `json:"status"`
}

func CreateCredential(key, token string, credits int) Credential {
	return Credential{
		Key:     key,
		Token:   token,
		Credits: credits,
		Status:  false,
	}
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configFilePath := filepath.Join(homeDir, configDir, configFileName)
	return configFilePath, nil
}

func write(configFilePath string, cfg Config) error {
	dir := filepath.Dir(configFilePath)
	if _, err := os.Stat(dir); err != nil && errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
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

func read(configFilePath string) (Config, error) {
	if _, err := os.Stat(configFilePath); err != nil && errors.Is(err, os.ErrNotExist) {
		write(configFilePath, Config{Credentials: map[string]Credential{}})
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

func Read() (Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	cfg, err := read(configFilePath)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) addCredential(configFilePath, email string, credentials Credential) error {
	if len(c.Credentials) == 0 {
		credentials.Status = true
	}

	c.Credentials[email] = credentials
	return write(configFilePath, *c)
}

func (c *Config) AddCredential(email string, credentials Credential) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	return c.addCredential(configFilePath, email, credentials)
}

// func (c *Config) GetKeyInfo() KeyInfo {
// 	var foundKey string

// 	for key, value := range c.Credentials {
// 		if value.Status {
// 			foundKey = key
// 			break
// 		}
// 	}

// 	c.SetStatus(activeKey)
// 	return c.APIKeys[activeKey]
// }

// func (c *Config) SetStatus(activeKey int) error {
// 	for index := range c.APIKeys {
// 		if index == activeKey {
// 			c.APIKeys[index].Status = true
// 		} else {
// 			c.APIKeys[index].Status = false
// 		}
// 	}
// 	return write(*c)
// }

// func (c *Config) SetToken(key string, token string) error {
// 	for i, k := range c.APIKeys {
// 		if k.Key == key {
// 			c.APIKeys[i].Token = token
// 			break
// 		}
// 	}

// 	return write(*c)
// }

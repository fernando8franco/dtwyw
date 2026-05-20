package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetConfigFilePath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, configDir, configFileName)

	got, err := getConfigFilePath()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), configFileName)
	expected := Config{
		Credentials: map[string]Credential{
			"test@test.com": {
				Key:     "api-key-123",
				Token:   "token",
				Credits: 100,
				Status:  true,
			},
		},
	}

	err := write(path, expected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file was not created: %v", err)
	}

	var cfg Config
	json.Unmarshal(data, &cfg)
	if !reflect.DeepEqual(expected, cfg) {
		t.Errorf("Returned config does not match expected. Got: %+v, Want: %+v", cfg, expected)
	}
}

func TestRead_FileNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), configFileName)
	expected := Config{Credentials: map[string]Credential{}}

	cfg, err := read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expected, cfg) {
		t.Errorf("Returned config does not match expected. Got: %+v, Want: %+v", cfg, expected)
	}
}

func TestRead_FileExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), configFileName)
	expected := Config{
		Credentials: map[string]Credential{
			"test@test.com": {
				Key:     "api-key-123",
				Token:   "token",
				Credits: 100,
				Status:  true,
			},
		},
	}

	write(path, expected)
	cfg, err := read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expected, cfg) {
		t.Errorf("Returned config does not match expected. Got: %+v, Want: %+v", cfg, expected)
	}
}

func TestAddCredential(t *testing.T) {
	path := filepath.Join(t.TempDir(), configFileName)
	cfg := Config{Credentials: map[string]Credential{}}

	cred := Credential{Key: "api-key-123", Token: "abc-123", Credits: 100}

	err := cfg.addCredential(path, "test@test.com", cred)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cred.Status = true
	expected := Config{Credentials: map[string]Credential{"test@test.com": cred}}

	if !reflect.DeepEqual(expected, cfg) {
		t.Errorf("Config mismatch.\nGot:  %+v\nWant: %+v", cfg, expected)
	}
}

func TestAddCredential_ExclusiveStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), configFileName)
	cfg := Config{
		Credentials: map[string]Credential{
			"existing@test.com": {
				Key:     "api-key-123",
				Token:   "token",
				Credits: 100,
				Status:  true,
			},
		},
	}

	newCred := Credential{Key: "api-key-456", Token: "abc-456", Credits: 200}
	err := cfg.addCredential(path, "new@test.com", newCred)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Credentials["new@test.com"].Status {
		t.Error("Expected new credential to have Status: false")
	}

	if !cfg.Credentials["existing@test.com"].Status {
		t.Error("Expected previous active credential to be set to Status: true")
	}
}

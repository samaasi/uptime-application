package seeder

import (
	"os"
	"testing"
)

func TestLoadSeedConfig(t *testing.T) {
	// Test loading embedded default config
	config, err := LoadSeedConfig("nonexistent.yaml")
	if err != nil {
		t.Fatalf("Expected to load default config when file doesn't exist, got error: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be loaded, got nil")
	}

	// Verify some basic structure
	if len(config.Permissions) == 0 {
		t.Error("Expected permissions to be loaded from default config")
	}

	if len(config.OrganizationTypes) == 0 {
		t.Error("Expected organization types to be loaded from default config")
	}

	if len(config.ApplicationTypes) == 0 {
		t.Error("Expected application types to be loaded from default config")
	}
}

func TestGetDefaultSeedConfig(t *testing.T) {
	config, err := getDefaultSeedConfig()
	if err != nil {
		t.Fatalf("Failed to get default seed config: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be loaded, got nil")
	}

	// Verify the config has expected data
	if len(config.Permissions) == 0 {
		t.Error("Expected permissions in default config")
	}

	// Check for some expected permissions
	foundUserRead := false
	for _, perm := range config.Permissions {
		if perm.Name == "user:read" {
			foundUserRead = true
			break
		}
	}
	if !foundUserRead {
		t.Error("Expected to find 'user:read' permission in default config")
	}
}

func TestSaveAndLoadSeedConfig(t *testing.T) {
	// Create a temporary file
	tmpFile := "test_config.yaml"
	defer os.Remove(tmpFile)

	// Get default config
	originalConfig, err := getDefaultSeedConfig()
	if err != nil {
		t.Fatalf("Failed to get default config: %v", err)
	}

	// Save config to file
	err = SaveSeedConfig(originalConfig, tmpFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config from file
	loadedConfig, err := LoadSeedConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	// Verify the loaded config matches the original
	if len(loadedConfig.Permissions) != len(originalConfig.Permissions) {
		t.Errorf("Expected %d permissions, got %d", len(originalConfig.Permissions), len(loadedConfig.Permissions))
	}

	if len(loadedConfig.OrganizationTypes) != len(originalConfig.OrganizationTypes) {
		t.Errorf("Expected %d organization types, got %d", len(originalConfig.OrganizationTypes), len(loadedConfig.OrganizationTypes))
	}

	if len(loadedConfig.ApplicationTypes) != len(originalConfig.ApplicationTypes) {
		t.Errorf("Expected %d application types, got %d", len(originalConfig.ApplicationTypes), len(loadedConfig.ApplicationTypes))
	}
}
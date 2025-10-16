package seeder

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/samaasi/uptime-application/services/api-services/internal/api/models"
	"github.com/samaasi/uptime-application/services/api-services/internal/utils"
	"gopkg.in/yaml.v3"
)

//go:embed seed_config.yaml
var defaultSeedConfigYAML []byte

// SeedConfig represents the configuration for seed data
type SeedConfig struct {
	Permissions       []PermissionConfig       `yaml:"permissions"`
	OrganizationTypes []OrganizationTypeConfig `yaml:"organization_types"`
	ApplicationTypes  []ApplicationTypeConfig  `yaml:"application_types"`
}

// PermissionConfig represents permission configuration
type PermissionConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// OrganizationTypeConfig represents organization type configuration
type OrganizationTypeConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// ApplicationTypeConfig represents application type configuration
type ApplicationTypeConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// ToModel converts PermissionConfig to models.Permission
func (pc PermissionConfig) ToModel() models.Permission {
	return models.Permission{
		Name:        pc.Name,
		Description: utils.StringPtr(pc.Description),
	}
}

// ToModel converts OrganizationTypeConfig to models.OrganizationType
func (otc OrganizationTypeConfig) ToModel() models.OrganizationType {
	return models.OrganizationType{
		Name:        otc.Name,
		Description: utils.StringPtr(otc.Description),
	}
}

// ToModel converts ApplicationTypeConfig to models.ApplicationType
func (atc ApplicationTypeConfig) ToModel() models.ApplicationType {
	return models.ApplicationType{
		Name:        atc.Name,
		Description: utils.StringPtr(atc.Description),
	}
}

// LoadSeedConfig loads seed configuration from the specified path
// If the file doesn't exist, it returns embedded default configuration
func LoadSeedConfig(configPath string) (*SeedConfig, error) {
	if envPath := os.Getenv("SEED_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	// If no path provided, use default
	if configPath == "" {
		configPath = "./seed_config.yaml"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultSeedConfig()
	}

	// Read and parse the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config SeedConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return &config, nil
}

// SaveSeedConfig saves the current configuration to a file
func SaveSeedConfig(config *SeedConfig, configPath string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getDefaultSeedConfig returns the default seed configuration from embedded YAML
func getDefaultSeedConfig() (*SeedConfig, error) {
	var config SeedConfig
	if err := yaml.Unmarshal(defaultSeedConfigYAML, &config); err != nil {
		return nil, fmt.Errorf("failed to parse embedded seed config: %w", err)
	}
	return &config, nil
}

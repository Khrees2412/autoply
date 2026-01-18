package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	OpenAIKey     string `mapstructure:"openai_key"`
	AnthropicKey  string `mapstructure:"anthropic_key"`
	AIProvider    string `mapstructure:"ai_provider"` // openai, anthropic
	DefaultModel  string `mapstructure:"default_model"`
}

var AppConfig *Config

// Initialize loads or creates the configuration file
func Initialize() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".autoply")
	configFile := filepath.Join(configDir, "config.yaml")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create default config if it doesn't exist
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := createDefaultConfig(configFile); err != nil {
			return err
		}
	}

	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	// Set defaults
	viper.SetDefault("ai_provider", "openai")
	viper.SetDefault("default_model", "gpt-4")
	viper.SetDefault("openai_key", "")
	viper.SetDefault("anthropic_key", "")

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Unmarshal into struct
	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// createDefaultConfig creates a default config file
func createDefaultConfig(path string) error {
	defaultConfig := `# Autoply Configuration
# AI Provider: openai or anthropic
ai_provider: openai
default_model: gpt-4

# API Keys (keep this file secure!)
openai_key: ""
anthropic_key: ""
`
	return os.WriteFile(path, []byte(defaultConfig), 0600)
}

// Set updates a configuration value
func Set(key, value string) error {
	viper.Set(key, value)
	return viper.WriteConfig()
}

// Get retrieves a configuration value
func Get(key string) string {
	return viper.GetString(key)
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".autoply", "config.yaml")
}

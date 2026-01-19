package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	OpenAIKey    string `mapstructure:"openai_key"`
	AnthropicKey string `mapstructure:"anthropic_key"`
	AIProvider   string `mapstructure:"ai_provider"` // openai, anthropic, ollama, lmstudio
	DefaultModel string `mapstructure:"default_model"`
	OllamaURL    string `mapstructure:"ollama_url"`
	LMStudioURL  string `mapstructure:"lmstudio_url"`
	// Job board credentials
	LinkedInEmail    string `mapstructure:"linkedin_email"`
	LinkedInPassword string `mapstructure:"linkedin_password"`
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
	viper.SetDefault("ai_provider", "ollama")
	viper.SetDefault("default_model", "llama3.2")
	viper.SetDefault("ollama_url", "http://localhost:11434")
	viper.SetDefault("lmstudio_url", "http://localhost:1234")
	viper.SetDefault("openai_key", "")
	viper.SetDefault("anthropic_key", "")
	viper.SetDefault("linkedin_email", "")
	viper.SetDefault("linkedin_password", "")

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
# AI Provider: openai, anthropic, ollama, lmstudio
ai_provider: ollama
default_model: llama3.2
ollama_url: http://localhost:11434
lmstudio_url: http://localhost:1234

# API Keys (keep this file secure!)
openai_key: ""
anthropic_key: ""

# Job Board Credentials (keep this file secure!)
linkedin_email: ""
linkedin_password: ""
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

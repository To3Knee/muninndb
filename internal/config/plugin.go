package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const pluginConfigFile = "plugin_config.json"

// PluginConfig holds persistent plugin configuration saved to disk.
// Environment variables always take precedence over values stored here.
type PluginConfig struct {
	// Embed provider settings
	EmbedProvider string `json:"embed_provider"` // "ollama", "openai", "voyage", "local", "none"
	EmbedURL      string `json:"embed_url"`      // provider URL (used for ollama)
	EmbedAPIKey   string `json:"embed_api_key"`  // API key (openai, voyage)

	// Enrich provider settings
	EnrichProvider string `json:"enrich_provider"` // "ollama", "openai", "anthropic"
	EnrichURL      string `json:"enrich_url"`      // full provider URL
	EnrichAPIKey   string `json:"enrich_api_key"`  // API key
}

// LoadPluginConfig reads plugin_config.json from dataDir.
// Returns an empty PluginConfig (not an error) if the file does not exist.
func LoadPluginConfig(dataDir string) (PluginConfig, error) {
	path := filepath.Join(dataDir, pluginConfigFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return PluginConfig{}, nil
	}
	if err != nil {
		return PluginConfig{}, err
	}
	var cfg PluginConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return PluginConfig{}, err
	}
	return cfg, nil
}

// SavePluginConfig writes cfg to plugin_config.json in dataDir.
func SavePluginConfig(dataDir string, cfg PluginConfig) error {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, pluginConfigFile), data, 0600)
}

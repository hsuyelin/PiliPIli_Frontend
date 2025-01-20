package config

import (
	"PiliPili_Frontend/util"
	"github.com/spf13/viper"
)

// Config holds all configuration values.
type Config struct {
	LogLevel               string               // Log level (e.g., INFO, DEBUG, ERROR)
	Encipher               string               // Key used for encryption and obfuscation
	EmbyURL                string               // Emby server URL
	EmbyPort               int                  // Emby server port
	EmbyAPIKey             string               // API key for Emby server
	BackendURL             string               // Backend streaming server URL
	BackendStorageBasePath string               // Backend streaming storage base path
	PlayURLMaxAliveTime    int                  // Maximum lifetime of the play URL
	ServerPort             int                  // Server port
	SpecialMedias          []SpecialMediaConfig // Special media configurations as a list
}

// SpecialMediaConfig holds the media path and source ID for a specific media.
type SpecialMediaConfig struct {
	Key           string // Unique key for the special media
	Name          string // Description of the special media
	MediaPath     string // Path to the media file
	ItemId        string // Item ID
	MediaSourceID string // Media source ID
}

// globalConfig stores the loaded configuration.
var globalConfig Config

// Initialize loads the configuration from the provided config file and initializes the logger.
func Initialize(configFile string, loglevel string) error {
	viper.SetConfigType("yaml")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		// Default configuration
		globalConfig = Config{
			LogLevel:               defaultLogLevel(loglevel),
			Encipher:               "vPQC5LWCN2CW2opz",
			EmbyURL:                "http://127.0.0.1",
			EmbyPort:               8096,
			EmbyAPIKey:             "",
			BackendURL:             "",
			BackendStorageBasePath: "",
			PlayURLMaxAliveTime:    6 * 60 * 60,
			ServerPort:             60002,
			SpecialMedias:          []SpecialMediaConfig{},
		}
	} else {
		// Load configuration from file
		globalConfig = Config{
			LogLevel:               getLogLevel(loglevel),
			Encipher:               viper.GetString("Encipher"),
			EmbyURL:                viper.GetString("Emby.url"),
			EmbyPort:               viper.GetInt("Emby.port"),
			EmbyAPIKey:             viper.GetString("Emby.apiKey"),
			BackendURL:             viper.GetString("Backend.url"),
			BackendStorageBasePath: viper.GetString("Backend.storageBasePath"),
			PlayURLMaxAliveTime:    viper.GetInt("PlayURLMaxAliveTime"),
			ServerPort:             viper.GetInt("Server.port"),
			SpecialMedias:          loadSpecialMedias(),
		}
	}

	return nil
}

// loadSpecialMedias parses the SpecialMedias configuration from viper.
func loadSpecialMedias() []SpecialMediaConfig {
	var specialMedias []SpecialMediaConfig

	if err := viper.UnmarshalKey("SpecialMedias", &specialMedias); err != nil {
		return []SpecialMediaConfig{}
	}

	return specialMedias
}

// GetConfig returns the global configuration.
func GetConfig() Config {
	return globalConfig
}

// IsValid checks if all fields in SpecialMediaConfig are non-empty and valid.
func (config SpecialMediaConfig) IsValid() bool {
	return config.Key != "" &&
		config.Name != "" &&
		config.MediaPath != "" &&
		config.ItemId != "" &&
		config.MediaSourceID != ""
}

// GetFullEmbyURL returns the complete Emby URL with the configured port.
func GetFullEmbyURL() string {
	return util.BuildFullURL(globalConfig.EmbyURL, globalConfig.EmbyPort)
}

// GetFullBackendURL returns the complete Backend URL.
func GetFullBackendURL() string {
	return util.BuildFullURL(globalConfig.BackendURL, 0)
}

// defaultLogLevel returns the default log level if no log level is specified.
func defaultLogLevel(loglevel string) string {
	if loglevel != "" {
		return loglevel
	}
	return "INFO"
}

// getLogLevel returns the log level from either the parameter or the config file.
func getLogLevel(loglevel string) string {
	if loglevel != "" {
		return loglevel
	}
	return viper.GetString("LogLevel")
}

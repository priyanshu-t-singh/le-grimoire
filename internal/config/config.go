package config

import (
	"errors"
	"fmt"
	"le-grimoire/internal/constants"
	"le-grimoire/internal/util"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	Version    string
	AppDataDir string
	Server     struct {
		Host string
		Port int
	}
	Logs struct {
		Dir string
	}
	Database struct {
		Name string
	}

	BookBackend  string // "kavita" | "localfs"
	KavitaURL    string
	KavitaAPIKey string
}

type ConfigOptions struct {
	Flags LeGrimoireFlags
}

func NewConfig(options *ConfigOptions, logger *slog.Logger) (*Config, error) {
	flags := options.Flags

	logger.Info("Loading configuration...")

	definedDataDir := ""

	// Set data dir (flag overrides env var)
	if os.Getenv("LE_GRIMOIRE_DATA_DIR") != "" {
		definedDataDir = os.Getenv("LE_GRIMOIRE_DATA_DIR")
	}

	if flags.DataDir != "" {
		definedDataDir = flags.DataDir
	}

	bookBackend := "kavita" // default backend
	if os.Getenv("BOOK_BACKEND") != "" {
		bookBackend = os.Getenv("BOOK_BACKEND")
	}

	defaultHost := constants.DefaultHost
	defaultPort := constants.DefaultPort

	// Environment variable overrides defaults
	if os.Getenv("LE_GRIMOIRE_SERVER_HOST") != "" {
		defaultHost = os.Getenv("LE_GRIMOIRE_SERVER_HOST")
	}
	if os.Getenv("LE_GRIMOIRE_SERVER_PORT") != "" {
		var err error
		defaultPort, err = strconv.Atoi(os.Getenv("LE_GRIMOIRE_SERVER_PORT"))
		if err != nil {
			return nil, fmt.Errorf("invalid LE_GRIMOIRE_SERVER_PORT environment variable: %s", os.Getenv("LE_GRIMOIRE_SERVER_PORT"))
		}
	}

	// Flag overrides environment variable
	if flags.Host != "" {
		defaultHost = flags.Host
	}
	if flags.Port != 0 {
		defaultPort = flags.Port
	}

	// Initialize the app data directory
	dataDir, configPath, err := initAppDataDir(definedDataDir, logger)
	if err != nil {
		return nil, err
	}

	if err = setDataDirEnv(dataDir); err != nil {
		return nil, err
	}

	// Configure viper
	viper.SetConfigName(constants.ConfigFileName)
	viper.SetConfigType("toml")
	viper.SetConfigFile(configPath)

	// Set default values
	viper.SetDefault("version", constants.Version)
	viper.SetDefault("server.host", defaultHost)
	viper.SetDefault("server.port", defaultPort)
	viper.SetDefault("database.name", "le-grimoire")
	viper.SetDefault("logs.dir", "$LE_GRIMOIRE_DATA_DIR/logs")
	viper.SetDefault("bookbackend", bookBackend)

	// Create and populate the config file if it doesn't exist
	if err := createConfigFile(configPath); err != nil {
		return nil, err
	}

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Check if host or port have been overridden and differ from config file
	existingHost := viper.GetString("server.host")
	existingPort := viper.GetInt("server.port")
	isHostChanged := false
	isPortChanged := false

	if (flags.Host != "" || os.Getenv("LE_GRIMOIRE_SERVER_HOST") != "") && existingHost != defaultHost {
		viper.Set("server.host", defaultHost)
		isHostChanged = true
	}
	if (flags.Port != 0 || os.Getenv("LE_GRIMOIRE_SERVER_PORT") != "") && existingPort != defaultPort {
		viper.Set("server.port", defaultPort)
		isPortChanged = true
	}

	// Write config if host or port have changed
	if isHostChanged || isPortChanged {
		if err := viper.WriteConfig(); err != nil {
			logger.Warn("Failed to write updated config with new host/port", "error", err)
		} else {
			logger.Info("Updated config with new host/port", "host", defaultHost, "port", defaultPort)
		}
	}

	// Unmarshal the config values
	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	expandEnvironmentValues(cfg)
	cfg.AppDataDir = dataDir

	return cfg, nil
}

func (cfg *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
}

func (cfg *Config) GetServerURI() string {
	return fmt.Sprintf("http://%s", cfg.GetServerAddr())
}

// TODO: remove this func
func (cfg *Config) GetKavitaAPIURI() string {
	return util.GetKavitaAPIKey()
}

func initAppDataDir(definedDataDir string, logger *slog.Logger) (dataDir string, configPath string, err error) {

	// User defined data dir
	if definedDataDir != "" {
		definedDataDir = filepath.FromSlash(os.ExpandEnv(definedDataDir))

		if !filepath.IsAbs(definedDataDir) {
			return "", "", errors.New("data directory path must be absolute")
		}

		// Replace the default data directory
		dataDir = definedDataDir

		logger.Debug("Overriding default data directory", slog.String("dataDir", dataDir))
	} else {
		dataDir, err = os.UserConfigDir()
		if err != nil {
			return "", "", err
		}

		dataDir = filepath.Join(dataDir, "le-grimoire")
	}

	// Create the data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return "", "", err
	}

	configPath = filepath.FromSlash(filepath.Join(dataDir, constants.ConfigFileName))
	dataDir = filepath.FromSlash(dataDir)

	logger.Debug("Using data directory", slog.String("dataDir", dataDir))
	logger.Debug("Using config path", slog.String("configPath", configPath))

	return dataDir, configPath, nil
}

func setDataDirEnv(dataDir string) error {
	// Set the data directory environment variable
	if os.Getenv("LE_GRIMOIRE_DATA_DIR") == "" {
		if err := os.Setenv("LE_GRIMOIRE_DATA_DIR", dataDir); err != nil {
			return err
		}
	}

	return nil
}

// creates a default config file if it doesn't exist
func createConfigFile(configPath string) error {
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
			return err
		}
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}
	return nil
}

func expandEnvironmentValues(cfg *Config) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic while expanding environment values:", r)
		}
	}()

	cfg.Logs.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Logs.Dir))
}

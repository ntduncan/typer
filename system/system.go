package system

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"ntduncan.com/typer/utils"
)

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("error getting home directory: %v", err))
	}

	// Build proper paths
	appDir := filepath.Join(homeDir, ".config", "funkeytype")
	configPath := filepath.Join(appDir, "config.json")

	return configPath
}

func checkIfDirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func SaveConfig(config Config) error {

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(getConfigPath(), jsonData, 0644)
	if err != nil {
		return fmt.Errorf("Error writing config to file: %s", err)
	}

	return nil
}

// Creates new .config/funkeytype path
// Initialize new .config/funkeytype/config.json?
func initConfig() error {
	defaultConfig := Config{
		Size:     10,
		Mode:     utils.WordsTest,
		TopScore: "0.00",
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %v", err)
	}

	// Build proper paths
	configDir := filepath.Join(homeDir, ".config")
	appDir := filepath.Join(homeDir, ".config", "funkeytype")
	configPath := filepath.Join(appDir, "config.json") // Assuming config filename

	// Create .config directory if it doesn't exist
	if !checkIfDirExists(configDir) {
		err = os.MkdirAll(configDir, 0755) // Use 0755 instead of fs.ModeDir
		if err != nil {
			return fmt.Errorf("error creating config directory: %v", err)
		}
	}

	// Create app directory if it doesn't exist
	if !checkIfDirExists(appDir) {
		err = os.MkdirAll(appDir, 0755) // Use 0755 and fix typo
		if err != nil {
			return fmt.Errorf("error creating app directory: %v", err)
		}
	}

	// Create and write config file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("error creating config file: %v", err)
	}
	defer file.Close()

	// Encode the struct directly, not the JSON bytes
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")         // Optional: pretty print
	err = encoder.Encode(defaultConfig) // Encode the struct, not jsonData
	if err != nil {
		return fmt.Errorf("error encoding config: %v", err)
	}

	return nil

}

func LoadConfig() (Config, error) {
	var config Config
	//Read COnfig
	_, err := os.Stat(getConfigPath())

	//IF no config, initConfig
	if errors.Is(err, os.ErrNotExist) {
		err = initConfig()
		if err != nil {
			return config, err
		}
	}

	file, err := os.Open(getConfigPath())
	bytes, _ := io.ReadAll(file)

	defer file.Close()

	err = json.Unmarshal(bytes, &config)

	return config, err
}

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

func configPath() (string, error) {
	// Get the current user's home directory
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	homeDir := usr.HomeDir

	// Construct the path to the configuration file
	configPath := filepath.Join(homeDir, ".goget", "config.json")

	return configPath, nil
}

type Config struct {
	GitToken string `json:"git_token"`
}

func readConfig(filename string) (*Config, error) {
	// Read the file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func writeConfig(config *Config, cfgPath string) error {
	// Create the necessary directories
	err := os.MkdirAll(filepath.Dir(cfgPath), 0700)
	if err != nil {
		return err
	}

	// Marshal the config data to JSON
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	// Write the config data to the file
	err = ioutil.WriteFile(cfgPath, data, 0600)
	if err != nil {
		log.Fatalln("Did not make file")
		return err
	}

	// Set the permissions on the file so that it is only readable and writable by the current user
	err = os.Chmod(cfgPath, 0600)
	if err != nil {
		return err
	}

	return nil
}

func ReadGitToken() (string, error) {
	cfgPath, err := configPath()
	if err != nil {
		return "", err
	}

	cfg, err := readConfig(cfgPath)
	if err != nil {
		return "", err
	}
	return cfg.GitToken, nil
}

func WriteGitToken(token string) error {

	cfgPath, err := configPath()
	if err != nil {
		return err
	}

	cfg, err := readConfig(cfgPath)
	if err != nil {
		return writeConfig(&Config{
			GitToken: token,
		}, cfgPath)
	}

	cfg.GitToken = token
	return writeConfig(cfg, cfgPath)
}

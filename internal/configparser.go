package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Plugin struct {
	RepositoryUrl string `json:"repositoryUrl,omitempty"`
	DownloadUrl   string `json:"downloadUrl,omitempty"`
	Credentials   string `json:"credentials,omitempty"`
	Artifact      string `json:"artifact,omitempty"`
}

type Config struct {
	Credentials map[string]Credentials `json:"credentials"`
	Plugins     []Plugin               `json:"plugins"`
}

func Parse(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	// Replace $ prefix strings in credential fields with environment variables
	for key, creds := range config.Credentials {
		creds.Username, err = replaceEnvVar(creds.Username)
		if err != nil {
			panic(err)
		}

		creds.Password, err = replaceEnvVar(creds.Password)
		if err != nil {
			panic(err)
		}

		config.Credentials[key] = creds
	}

	fmt.Println("Credentials:")
	for key, creds := range config.Credentials {
		fmt.Printf("%s: Username: %s, Password: %s\n", key, creds.Username, creds.Password)
	}

	fmt.Println("\nPlugins:")
	for _, plugin := range config.Plugins {
		fmt.Println("Repository URL:", plugin.RepositoryUrl)
		fmt.Println("Download URL:", plugin.DownloadUrl)
		fmt.Println("Credentials:", plugin.Credentials)
		fmt.Println("Artifact:", plugin.Artifact)
		fmt.Println()
	}
}

func replaceEnvVar(s string) (string, error) {
	if strings.HasPrefix(s, "$") {
		envVar := strings.TrimPrefix(s, "$")
		if value, ok := os.LookupEnv(envVar); ok {
			return value, nil
		}
		return "", fmt.Errorf("environment variable not found: %s", envVar)
	}
	return s, nil
}

package internal

import (
	"encoding/json"
	"fmt"
	"log"
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

func (p Plugin) GetDownloadURL() string {
	if p.DownloadUrl != "" {
		return p.DownloadUrl
	}
	if p.RepositoryUrl != "" && p.Artifact != "" {
		parts := strings.Split(p.Artifact, ":")
		if len(parts) != 3 {
			return "" // invalid format
		}
		groupId := strings.ReplaceAll(parts[0], ".", "/")
		artifactId := parts[1]
		version := parts[2]

		return fmt.Sprintf("%s/%s/%s/%s/%s-%s.jar", p.RepositoryUrl, groupId, artifactId, version, artifactId, version)
	}
	return ""
}

type Config struct {
	Credentials map[string]Credentials `json:"credentials"`
	Plugins     []Plugin               `json:"plugins"`
}

func Parse(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Error opening JSON file: ", err)
		return
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		log.Fatal("Error decoding JSON: ", err)
		return
	}

	// Replace $ prefix strings in credential fields with environment variables
	for key, creds := range config.Credentials {
		creds.Username, err = replaceEnvVar(creds.Username)
		if err != nil {
			log.Fatal(err)
		}

		creds.Password, err = replaceEnvVar(creds.Password)
		if err != nil {
			log.Fatal(err)
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
		fmt.Println("download:", plugin.GetDownloadURL())
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

package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
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

func (p Plugin) Filename() string {
	return filepath.Base(p.GetDownloadURL())
}

type Config struct {
	Credentials map[string]Credentials `json:"credentials"`
	Plugins     []Plugin               `json:"plugins"`
}

func Parse(filePath string, logger *zap.Logger) *Config {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Fatal("Error opening JSON file", zap.Error(err))
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		logger.Fatal("Error decoding JSON", zap.Error(err))
	}

	// Replace $ prefix strings in credential fields with environment variables
	for key, creds := range config.Credentials {
		creds.Username, err = replaceEnvVar(creds.Username)
		if err != nil {
			logger.Fatal("Failed to replace username from env", zap.Error(err))
		}

		creds.Password, err = replaceEnvVar(creds.Password)
		if err != nil {
			logger.Fatal("Failed to replace password from env", zap.Error(err))
		}

		config.Credentials[key] = creds
	}

	return &config
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

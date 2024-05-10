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

type DownloadableItem struct {
	Credentials     string   `json:"credentials,omitempty"`
	AddToLocalMaven bool     `json:"addToLocalMaven,omitempty"`
	Tags            []string `json:"tags,omitempty"`

	DownloadUrl string `json:"downloadUrl,omitempty"`
	// should not be used with repositoryUrl
	SaveAs string `json:"saveAs,omitempty"`

	RepositoryUrl string `json:"repositoryUrl,omitempty"`
	ArtifactID    string `json:"artifactId,omitempty"`
	GroupID       string `json:"groupId,omitempty"`
	Version       string `json:"version,omitempty"`
}

type LocalMavenConfig struct {
	GroupId    string `json:"groupId,omitempty"`
	ArtifactId string `json:"artifactId,omitempty"`
	Version    string `json:"version,omitempty"`
}

func (p DownloadableItem) GetDownloadURL() string {
	if p.DownloadUrl != "" {
		return p.DownloadUrl
	}

	if p.RepositoryUrl != "" && p.GroupID != "" && p.ArtifactID != "" && p.Version != "" {
		return fmt.Sprintf("%s/%s/%s/%s/%s-%s.jar", p.RepositoryUrl, p.GroupID, p.ArtifactID, p.Version, p.ArtifactID, p.Version)
	}

	return ""
}

func (p DownloadableItem) Filename() string {
	// sometimes the download url may end with /download so it won't have the .jar extension and stuff like that
	if p.SaveAs != "" {
		return p.SaveAs
	}

	return filepath.Base(p.GetDownloadURL())
}

type Config struct {
	Credentials map[string]Credentials `json:"credentials"`
	Plugins     []DownloadableItem     `json:"plugins"`
	Libraries   []DownloadableItem     `json:"libraries"` // the only difference between plugins and libs is that plugins array will be copied to the final jar to be downloaded later
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

func GetPlugins(config *Config) []DownloadableItem {
	pl := config.Plugins

	for _, lib := range config.Libraries {
		for _, tag := range lib.Tags {
			if tag == "plugin" {
				pl = append(pl, lib)
			}
		}
	}

	return pl
}

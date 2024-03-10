package internal

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"
)

func Download(config *Config, outdir string, logger *zap.Logger) []error {
	var errs []error

	for _, value := range config.Plugins {
		err := handlePlugin(value, config, outdir, logger)

		if err != nil {
			errs = append(errs,
				fmt.Errorf("failed to download (%s): %w", value.GetDownloadURL(), err))
		}

		continue
	}

	return errs
}

func handlePlugin(plugin Plugin, config *Config, outdir string, logger *zap.Logger) error {
	if plugin.GetDownloadURL() == "" {
		return nil
	}

	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", plugin.GetDownloadURL(), nil)

	if err != nil {
		return fmt.Errorf("failed to create new http client: %w", err)
	}

	if plugin.Credentials != "" {
		creds, ok := config.Credentials[plugin.Credentials]

		if !ok {
			return fmt.Errorf("invalid credentials reference for plugin: %s", plugin.GetDownloadURL())
		}

		req.SetBasicAuth(creds.Username, creds.Password)
	}

	logger.Info("Downloading plugin", zap.String("url", plugin.GetDownloadURL()))
	response, err := httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("failed to download object: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	err = SaveContentToFile(plugin.Filename(), response.Body, outdir, logger)

	if err != nil {
		logger.Error("Failed to save downloaded content into file", zap.Error(err))
	}

	if plugin.AddToLocalMaven {
		filePath := filepath.Join(outdir, plugin.Filename())

		cmd := exec.Command("mvn", "install:install-file",
			fmt.Sprintf("-Dfile=%s", filePath),
			fmt.Sprintf("-DgroupId=%s", plugin.LocalMavenConfig.GroupId),
			fmt.Sprintf("-DartifactId=%s", plugin.LocalMavenConfig.ArtifactId),
			fmt.Sprintf("-Dversion=%s", plugin.LocalMavenConfig.Version),
			"-Dpackaging=jar",
		)

		out, err := cmd.Output()

		if err != nil {
			logger.Error("Failed to add plugin to local maven repository", zap.Error(err))
			log.Default().Print(string(out))
		} else {
			log.Default().Print(string(out))
			logger.Info("Added plugin to local repository")
		}
	}

	return nil
}

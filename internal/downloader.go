package internal

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

func Download(items []DownloadableItem, credentials map[string]Credentials, outdir string, logger *zap.Logger, tags []string, m2RepoPath string) []error {
	var errs []error

	checkTags := len(tags) != 0

	if checkTags {
		logger.Info("Filtering based on tags", zap.String("tags", strings.Join(tags, ",")))
	}

	for _, value := range items {
		if checkTags && !commonTags(tags, value.Tags) {
			logger.Info("Skipping item, not included in tags", zap.String("item", value.Filename()), zap.String("tags", strings.Join(value.Tags, ",")))
			continue
		}

		err := handleItem(value, credentials, outdir, logger, m2RepoPath)

		if err != nil {
			errs = append(errs,
				fmt.Errorf("failed to download (%s): %w", value.GetDownloadURL(), err))
		}

		continue
	}

	return errs
}

func handleItem(item DownloadableItem, credentials map[string]Credentials, outdir string, logger *zap.Logger, m2RepoPath string) error {
	if item.GetDownloadURL() == "" {
		return nil
	}

	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", item.GetDownloadURL(), nil)

	if err != nil {
		return fmt.Errorf("failed to create new http client: %w", err)
	}

	if item.Credentials != "" {
		creds, ok := credentials[item.Credentials]

		if !ok {
			return fmt.Errorf("invalid credentials reference for item: %s", item.GetDownloadURL())
		}

		req.SetBasicAuth(creds.Username, creds.Password)
	}

	logger.Info("Downloading item", zap.String("url", item.GetDownloadURL()))
	response, err := httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("failed to download object: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code (%s): %d", item.GetDownloadURL(), response.StatusCode)
	}

	err = SaveContentToFile(item.Filename(), response.Body, outdir, logger)

	if err != nil {
		logger.Error("Failed to save downloaded content into file", zap.Error(err))
	}

	if item.AddToLocalMaven {
		filePath := filepath.Join(outdir, item.Filename())

		cmd := exec.Command("mvn", "install:install-file",
			fmt.Sprintf("-Dfile=%s", filePath),
			fmt.Sprintf("-DgroupId=%s", item.GroupID),
			fmt.Sprintf("-DartifactId=%s", item.ArtifactID),
			fmt.Sprintf("-Dversion=%s", item.Version),
			"-Dpackaging=jar",
			fmt.Sprintf("-DlocalRepositoryPath=%s", m2RepoPath),
		)

		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		err := cmd.Run()

		logger.Debug("maven logs", zap.String("stdout", out.String()), zap.String("stderr", stderr.String()))

		if err != nil {
			return fmt.Errorf("failed to add item to local maven repository: %s", err)
		} else {
			logger.Info("Added item to local repository", zap.String("name", item.Filename()))
		}
	}

	return nil
}

func commonTags(filterTags, tagsToCheck []string) bool {
	for _, filterTag := range filterTags {
		for _, tagToCheck := range tagsToCheck {
			if filterTag == tagToCheck {
				return true
			}
		}
	}

	return false
}

package internal

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"

	"go.uber.org/zap"
)

func extractDependenciesJSONFromJar(logger *zap.Logger, path string) (*Config, error) {
	fileToExtract := "dependencies.json"

	readCloser, err := extractFileFromJar(path, fileToExtract)
	if err != nil {
		if err.Error() == "file dependencies.json not found in JAR" {
			return nil, nil
		}

		logger.Error("Error extracting 'dependencies.json'", zap.String("jar", path), zap.Error(err))
		return nil, err
	}

	var conf Config
	err = json.NewDecoder(readCloser).Decode(&conf)
	defer readCloser.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to decode json from readcloser: %s", err)
	}

	return &conf, nil
}

func extractFileFromJar(jarPath, fileToExtract string) (io.ReadCloser, error) {
	// Open the JAR file
	jarFile, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, err
	}

	// Find the file to extract in the JAR
	var foundFile *zip.File
	for _, f := range jarFile.File {
		if f.Name == fileToExtract {
			foundFile = f
			break
		}
	}

	if foundFile == nil {
		return nil, fmt.Errorf("file %s not found in JAR", fileToExtract)
	}

	// Open the file inside the JAR
	zipFile, err := foundFile.Open()
	if err != nil {
		return nil, err
	}

	return zipFile, nil
}

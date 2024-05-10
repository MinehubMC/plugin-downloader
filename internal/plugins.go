package internal

import (
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
)

// [groupID+artifactID]: version
type DownloadedPluginsMap map[string]string

func DownloadPlugins(plugins []DownloadableItem, credentials map[string]Credentials, outdir string, logger *zap.Logger, tags []string, m2RepoPath string) []error {
	downloadedPlugins := make(DownloadedPluginsMap)

	var errors []error

	logger.Info("Starting root level plugin download...")
	for _, plugin := range plugins {
		err := downloadPluginRecursive(plugin, downloadedPlugins, credentials, outdir, logger, m2RepoPath)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func downloadPluginRecursive(plugin DownloadableItem, pluginMap DownloadedPluginsMap, credentials map[string]Credentials, outdir string, logger *zap.Logger, m2RepoPath string) error {
	mapKey := plugin.GroupID + ":" + plugin.ArtifactID

	if version, ok := pluginMap[mapKey]; ok {
		if plugin.Version != version {
			return fmt.Errorf("version conflict (%s, %s) for %s", plugin.Version, version, mapKey)
		}

		return nil
	}

	// Download the plugin
	err := handleItem(plugin, credentials, outdir, logger, m2RepoPath)
	if err != nil {
		return err
	}

	pluginMap[mapKey] = plugin.Version

	//Get dependencies of the plugin
	conf, err := extractDependenciesJSONFromJar(logger, filepath.Join(outdir, plugin.Filename()))

	if err != nil {
		return err
	} else if conf == nil {
		return nil
	}

	dependencies := GetPlugins(conf)

	if len(dependencies) > 0 {
		logger.Info("Found dependencies, downloading...", zap.String("for", mapKey+":"+plugin.Version), zap.Int("count", len(dependencies)))
	}

	for _, dep := range dependencies {
		err := downloadPluginRecursive(dep, pluginMap, credentials, outdir, logger, m2RepoPath)
		if err != nil {
			return err
		}
	}

	return nil
}

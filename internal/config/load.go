package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/okarahan/repo-compliance-checker/internal/model"
)

// LoadRepos reads repos JSON from path and unmarshals it into model.ReposConfig.
func LoadRepos(path string) (model.ReposConfig, error) {
	var cfg model.ReposConfig
	if err := loadJSON(path, &cfg); err != nil {
		return model.ReposConfig{}, fmt.Errorf("load repos config: %w", err)
	}
	return cfg, nil
}

// LoadAllowedTechnologies reads allowed technologies JSON from path and unmarshals it into model.AllowedTechnologies.
func LoadAllowedTechnologies(path string) (model.AllowedTechnologies, error) {
	var cfg model.AllowedTechnologies
	if err := loadJSON(path, &cfg); err != nil {
		return model.AllowedTechnologies{}, fmt.Errorf("load allowed technologies config: %w", err)
	}
	return cfg, nil
}

// LoadManifestMap reads the manifest mapping JSON from path and unmarshals it into model.ManifestMapConfig.
func LoadManifestMap(path string) (model.ManifestMapConfig, error) {
	var cfg model.ManifestMapConfig
	if err := loadJSON(path, &cfg); err != nil {
		return model.ManifestMapConfig{}, fmt.Errorf("load manifest map config: %w", err)
	}
	return cfg, nil
}

func loadJSON(path string, dest any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file %q: %w", path, err)
	}
	if err := json.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}
	return nil
}

package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Station struct {
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Tags []string `json:"tags,omitempty"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "rad.io", "stations.json"), nil
}

func Load() ([]Station, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Station{}, nil
		}
		return nil, err
	}
	var stations []Station
	if err := json.Unmarshal(data, &stations); err != nil {
		return nil, err
	}
	return stations, nil
}

func Save(stations []Station) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(stations, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

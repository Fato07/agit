package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/fathindos/agit/internal/config"
)

// CheckCache stores the result of the last update check.
type CheckCache struct {
	LastCheck     time.Time `json:"last_check"`
	LatestVersion string    `json:"latest_version"`
	LatestURL     string    `json:"latest_url"`
}

// CachePath returns the path to the update check cache file.
func CachePath() (string, error) {
	dir, err := config.AgitDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "update-check.json"), nil
}

// LoadCache reads the update check cache from disk.
// Returns an empty cache if the file doesn't exist.
func LoadCache() (*CheckCache, error) {
	path, err := CachePath()
	if err != nil {
		return &CheckCache{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CheckCache{}, nil
		}
		return &CheckCache{}, err
	}

	var cache CheckCache
	if err := json.Unmarshal(data, &cache); err != nil {
		// Corrupt cache file — treat as empty
		return &CheckCache{}, nil
	}
	return &cache, nil
}

// SaveCache writes the update check cache to disk.
func SaveCache(cache *CheckCache) error {
	path, err := CachePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ShouldCheck returns true if enough time has elapsed since the last check.
func (c *CheckCache) ShouldCheck(interval time.Duration) bool {
	if c.LastCheck.IsZero() {
		return true
	}
	return time.Since(c.LastCheck) >= interval
}

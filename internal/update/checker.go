package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	releaseURL  = "https://api.github.com/repos/Fato07/agit/releases/latest"
	httpTimeout = 5 * time.Second
)

// ReleaseAsset represents a downloadable file attached to a GitHub release.
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ReleaseInfo represents a GitHub release.
type ReleaseInfo struct {
	TagName string         `json:"tag_name"`
	HTMLURL string         `json:"html_url"`
	Assets  []ReleaseAsset `json:"assets"`
}

// FetchLatestRelease queries the GitHub API for the latest agit release.
func FetchLatestRelease() (*ReleaseInfo, error) {
	client := &http.Client{Timeout: httpTimeout}

	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "agit-update-checker")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("could not decode release: %w", err)
	}

	return &release, nil
}

// IsNewer returns true if latestTag represents a newer version than current.
// Both values may optionally have a "v" prefix.
func IsNewer(current, latestTag string) bool {
	cur := parseVersion(current)
	lat := parseVersion(latestTag)
	if cur == nil || lat == nil {
		return false
	}

	for i := 0; i < 3; i++ {
		if lat[i] > cur[i] {
			return true
		}
		if lat[i] < cur[i] {
			return false
		}
	}
	return false
}

// parseVersion splits a version string like "v1.2.3" or "1.2.3" into [major, minor, patch].
func parseVersion(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return nil
	}

	nums := make([]int, 3)
	for i, p := range parts {
		// Strip any pre-release suffix (e.g., "3-rc1")
		if idx := strings.IndexAny(p, "-+"); idx >= 0 {
			p = p[:idx]
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}

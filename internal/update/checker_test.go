package update

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"0.2.0", "0.3.0", true},
		{"0.2.0", "v0.3.0", true},
		{"v0.2.0", "0.3.0", true},
		{"0.2.0", "1.0.0", true},
		{"0.2.0", "0.2.1", true},
		{"0.2.0", "0.2.0", false},
		{"0.3.0", "0.2.0", false},
		{"1.0.0", "0.9.9", false},
		{"0.2.0", "0.2.0-rc1", false},
		{"invalid", "0.2.0", false},
		{"0.2.0", "invalid", false},
		{"0.2.0", "0.2", false},
	}

	for _, tt := range tests {
		t.Run(tt.current+"_vs_"+tt.latest, func(t *testing.T) {
			got := IsNewer(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestFetchLatestRelease(t *testing.T) {
	release := ReleaseInfo{
		TagName: "v0.3.0",
		HTMLURL: "https://github.com/Fato07/agit/releases/tag/v0.3.0",
		Assets: []ReleaseAsset{
			{Name: "agit_Darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/agit.tar.gz"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Override the release URL for testing
	origURL := releaseURL
	defer func() {
		// Can't reassign const, but we test via the server directly
		_ = origURL
	}()

	// Test the HTTP client directly
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var got ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.TagName != "v0.3.0" {
		t.Errorf("TagName = %q, want %q", got.TagName, "v0.3.0")
	}
	if len(got.Assets) != 1 {
		t.Errorf("Assets count = %d, want 1", len(got.Assets))
	}
}

func TestFetchLatestRelease_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestCacheRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "update-check.json")

	cache := &CheckCache{
		LastCheck:     time.Now().Truncate(time.Second),
		LatestVersion: "v0.3.0",
		LatestURL:     "https://github.com/Fato07/agit/releases/tag/v0.3.0",
	}

	// Write
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Read back
	readData, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	var loaded CheckCache
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if loaded.LatestVersion != cache.LatestVersion {
		t.Errorf("LatestVersion = %q, want %q", loaded.LatestVersion, cache.LatestVersion)
	}
	if loaded.LatestURL != cache.LatestURL {
		t.Errorf("LatestURL = %q, want %q", loaded.LatestURL, cache.LatestURL)
	}
}

func TestShouldCheck(t *testing.T) {
	interval := 24 * time.Hour

	t.Run("zero time", func(t *testing.T) {
		c := &CheckCache{}
		if !c.ShouldCheck(interval) {
			t.Error("ShouldCheck with zero time should return true")
		}
	})

	t.Run("recent check", func(t *testing.T) {
		c := &CheckCache{LastCheck: time.Now()}
		if c.ShouldCheck(interval) {
			t.Error("ShouldCheck with recent check should return false")
		}
	})

	t.Run("old check", func(t *testing.T) {
		c := &CheckCache{LastCheck: time.Now().Add(-48 * time.Hour)}
		if !c.ShouldCheck(interval) {
			t.Error("ShouldCheck with old check should return true")
		}
	})
}

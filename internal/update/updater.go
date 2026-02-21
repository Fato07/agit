package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallMethod describes how agit was installed.
type InstallMethod int

const (
	InstallBinary InstallMethod = iota
	InstallGoPath
)

// DetectInstallMethod checks whether the running binary lives in GOPATH/bin.
func DetectInstallMethod() InstallMethod {
	exe, err := os.Executable()
	if err != nil {
		return InstallBinary
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return InstallBinary
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return InstallBinary
		}
		gopath = filepath.Join(home, "go")
	}
	gopathBin := filepath.Join(gopath, "bin")

	if strings.HasPrefix(exe, gopathBin) {
		return InstallGoPath
	}
	return InstallBinary
}

// SelfUpdate downloads and installs the latest release.
func SelfUpdate(release *ReleaseInfo) error {
	method := DetectInstallMethod()

	switch method {
	case InstallGoPath:
		return updateViaGoInstall()
	default:
		return updateViaBinaryDownload(release)
	}
}

func updateViaGoInstall() error {
	cmd := exec.Command("go", "install", "github.com/fathindos/agit@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}
	return nil
}

func updateViaBinaryDownload(release *ReleaseInfo) error {
	assetName := buildAssetName()
	checksumName := "checksums.txt"

	var assetURL, checksumURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			assetURL = a.BrowserDownloadURL
		}
		if a.Name == checksumName {
			checksumURL = a.BrowserDownloadURL
		}
	}

	if assetURL == "" {
		return fmt.Errorf("no release asset found for %s/%s (expected %s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	// Download the archive
	tmpDir, err := os.MkdirTemp("", "agit-update-*")
	if err != nil {
		return fmt.Errorf("could not create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(archivePath, assetURL); err != nil {
		return fmt.Errorf("could not download release: %w", err)
	}

	// Verify checksum if available
	if checksumURL != "" {
		checksumPath := filepath.Join(tmpDir, checksumName)
		if err := downloadFile(checksumPath, checksumURL); err == nil {
			if err := verifyChecksum(archivePath, checksumPath, assetName); err != nil {
				return fmt.Errorf("checksum verification failed: %w", err)
			}
		}
	}

	// Extract binary
	binPath := filepath.Join(tmpDir, "agit")
	if err := extractBinary(archivePath, binPath); err != nil {
		return fmt.Errorf("could not extract binary: %w", err)
	}

	// Replace current executable
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find current executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %w", err)
	}

	if err := atomicReplace(binPath, exe); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied — try: sudo agit update")
		}
		return fmt.Errorf("could not replace binary: %w", err)
	}

	return nil
}

func buildAssetName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Match GoReleaser naming convention
	archName := goarch
	switch goarch {
	case "amd64":
		archName = "x86_64"
	case "arm64":
		archName = "arm64"
	}

	osName := strings.ToUpper(goos[:1]) + goos[1:]
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}

	return fmt.Sprintf("agit_%s_%s.%s", osName, archName, ext)
}

func downloadFile(dst, url string) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func verifyChecksum(archivePath, checksumPath, assetName string) error {
	// Read checksums file
	data, err := os.ReadFile(checksumPath)
	if err != nil {
		return err
	}

	// Find expected hash
	var expectedHash string
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == assetName {
			expectedHash = parts[0]
			break
		}
	}
	if expectedHash == "" {
		return fmt.Errorf("no checksum found for %s", assetName)
	}

	// Compute actual hash
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actualHash := hex.EncodeToString(h.Sum(nil))

	if actualHash != expectedHash {
		return fmt.Errorf("expected %s, got %s", expectedHash, actualHash)
	}
	return nil
}

func extractBinary(archivePath, binPath string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, binPath)
	}
	return extractTarGz(archivePath, binPath)
}

func extractTarGz(archivePath, binPath string) error {
	cmd := exec.Command("tar", "xzf", archivePath, "-C", filepath.Dir(binPath), "agit")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractZip(archivePath, binPath string) error {
	cmd := exec.Command("unzip", "-o", "-j", archivePath, "agit.exe", "-d", filepath.Dir(binPath))
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Rename agit.exe to agit for consistency
	exePath := filepath.Join(filepath.Dir(binPath), "agit.exe")
	if _, err := os.Stat(exePath); err == nil {
		return os.Rename(exePath, binPath)
	}
	return nil
}

func atomicReplace(src, dst string) error {
	// Copy permissions from existing binary
	info, err := os.Stat(dst)
	if err != nil {
		return err
	}

	// Rename new binary into place (atomic on same filesystem)
	tmpDst := dst + ".new"
	if err := copyFile(src, tmpDst); err != nil {
		return err
	}
	if err := os.Chmod(tmpDst, info.Mode()); err != nil {
		os.Remove(tmpDst)
		return err
	}
	if err := os.Rename(tmpDst, dst); err != nil {
		os.Remove(tmpDst)
		return err
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

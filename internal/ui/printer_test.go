package ui

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("could not create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("could not read pipe: %v", err)
	}
	return string(out)
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("could not create pipe: %v", err)
	}
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("could not read pipe: %v", err)
	}
	return string(out)
}

func TestSuccess(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	out := captureStdout(t, func() {
		Success("done %s", "now")
	})
	if !strings.Contains(out, "done now") {
		t.Errorf("expected 'done now' in output, got %q", out)
	}
}

func TestWarning(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	out := captureStdout(t, func() {
		Warning("caution %d", 42)
	})
	if !strings.Contains(out, "caution 42") {
		t.Errorf("expected 'caution 42' in output, got %q", out)
	}
}

func TestErrorf(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	out := captureStderr(t, func() {
		Errorf("bad %s", "thing")
	})
	if !strings.Contains(out, "bad thing") {
		t.Errorf("expected 'bad thing' in output, got %q", out)
	}
}

func TestInfo(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	Quiet = false
	out := captureStdout(t, func() {
		Info("hello %s", "world")
	})
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected 'hello world' in output, got %q", out)
	}
}

func TestInfoQuiet(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	Quiet = true
	defer func() { Quiet = false }()

	out := captureStdout(t, func() {
		Info("should not appear")
	})
	if out != "" {
		t.Errorf("expected no output in quiet mode, got %q", out)
	}
}

func TestSection(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	out := captureStdout(t, func() {
		Section("test")
	})
	if !strings.Contains(out, "TEST") {
		t.Errorf("expected uppercase 'TEST' in output, got %q", out)
	}
}

func TestKeyValue(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	out := captureStdout(t, func() {
		KeyValue("Name", "agit")
	})
	if !strings.Contains(out, "Name") || !strings.Contains(out, "agit") {
		t.Errorf("expected 'Name' and 'agit' in output, got %q", out)
	}
}

func TestBullet(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	out := captureStdout(t, func() {
		Bullet("item %d", 1)
	})
	if !strings.Contains(out, "item 1") {
		t.Errorf("expected 'item 1' in output, got %q", out)
	}
}

func TestBanner(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	out := captureStdout(t, func() {
		Banner("1.0.0")
	})
	if !strings.Contains(out, "agit v1.0.0") {
		t.Errorf("expected 'agit v1.0.0' in output, got %q", out)
	}
}

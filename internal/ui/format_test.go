package ui

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestIsJSON(t *testing.T) {
	CurrentFormat = FormatJSON
	if !IsJSON() {
		t.Error("expected IsJSON true for FormatJSON")
	}

	CurrentFormat = FormatText
	if IsJSON() {
		t.Error("expected IsJSON false for FormatText")
	}
}

func TestRenderJSON(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("could not create pipe: %v", err)
	}
	os.Stdout = w

	data := map[string]string{"key": "value"}
	if err := RenderJSON(data); err != nil {
		t.Fatalf("RenderJSON error: %v", err)
	}

	w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("could not read pipe: %v", err)
	}

	s := string(out)
	if !strings.Contains(s, `"key"`) || !strings.Contains(s, `"value"`) {
		t.Errorf("expected JSON with key/value, got %q", s)
	}
}

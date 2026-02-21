package ui

import (
	"encoding/json"
	"fmt"
	"os"
)

// OutputFormat controls how commands render their results.
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

// CurrentFormat is the active output format, set during CLI initialization.
var CurrentFormat OutputFormat = FormatText

// IsJSON returns true when output should be JSON.
func IsJSON() bool {
	return CurrentFormat == FormatJSON
}

// RenderJSON marshals v as indented JSON and writes it to stdout.
func RenderJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("could not encode JSON: %w", err)
	}
	return nil
}

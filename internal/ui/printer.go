package ui

import (
	"fmt"
	"os"
	"strings"
)

// Quiet suppresses Info output when true.
var Quiet bool

// Success prints a green checkmark followed by a formatted message.
func Success(msg string, args ...interface{}) {
	fmt.Printf("%s %s\n", T.Success("\u2713"), fmt.Sprintf(msg, args...))
}

// Warning prints a yellow WARNING prefix followed by a formatted message.
func Warning(msg string, args ...interface{}) {
	fmt.Printf("%s %s\n", T.Warning("WARNING:"), fmt.Sprintf(msg, args...))
}

// Errorf prints a red error message to stderr.
func Errorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s\n", T.Error("ERROR:"), fmt.Sprintf(msg, args...))
}

// Info prints a cyan informational message; suppressed when Quiet is true.
func Info(msg string, args ...interface{}) {
	if Quiet {
		return
	}
	fmt.Printf("%s %s\n", T.Info("INFO:"), fmt.Sprintf(msg, args...))
}

// Section prints a bold uppercase section header.
func Section(title string) {
	fmt.Printf("\n%s\n", T.Bold(strings.ToUpper(title)))
}

// KeyValue prints an indented "Key: value" line with the value in muted color.
func KeyValue(key, value string) {
	fmt.Printf("  %s: %s\n", key, T.Muted(value))
}

// Blank prints an empty line.
func Blank() {
	fmt.Println()
}

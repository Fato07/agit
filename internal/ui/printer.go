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

// Warning prints a yellow warning symbol followed by a formatted message.
func Warning(msg string, args ...interface{}) {
	fmt.Printf("%s %s\n", T.Warning(Sym.Warning), fmt.Sprintf(msg, args...))
}

// Errorf prints a red error symbol and message to stderr.
func Errorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s\n", T.Error(Sym.Error), fmt.Sprintf(msg, args...))
}

// Info prints a cyan informational message; suppressed when Quiet is true.
func Info(msg string, args ...interface{}) {
	if Quiet {
		return
	}
	fmt.Printf("%s %s\n", T.Info(Sym.Info), fmt.Sprintf(msg, args...))
}

// Section prints a bold uppercase section header with an underline.
func Section(title string) {
	upper := strings.ToUpper(title)
	underline := strings.Repeat(Sym.BoxH, len(upper))
	fmt.Printf("\n%s\n%s\n", T.Bold(upper), T.Muted(underline))
}

// KeyValue prints an indented arrow-prefixed "Key: value" line.
func KeyValue(key, value string) {
	fmt.Printf("  %s %s: %s\n", T.Muted(Sym.Arrow), key, T.Muted(value))
}

// Bullet prints a bullet-prefixed indented message.
func Bullet(msg string, args ...interface{}) {
	fmt.Printf("  %s %s\n", Sym.Bullet, fmt.Sprintf(msg, args...))
}

// Banner prints the branded agit version string.
func Banner(version string) {
	fmt.Println(T.Brand(Sym.Zap + " agit v" + version))
}

// Blank prints an empty line.
func Blank() {
	fmt.Println()
}

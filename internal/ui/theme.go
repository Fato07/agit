package ui

import "github.com/fatih/color"

// Theme defines the color functions used throughout the CLI.
type Theme struct {
	Success func(a ...interface{}) string
	Warning func(a ...interface{}) string
	Error   func(a ...interface{}) string
	Info    func(a ...interface{}) string
	Muted   func(a ...interface{}) string
	Bold    func(a ...interface{}) string
	Accent  func(a ...interface{}) string
	Brand   func(a ...interface{}) string
}

// DefaultTheme uses the same colors agit has always used.
var DefaultTheme = Theme{
	Success: color.New(color.FgGreen).SprintFunc(),
	Warning: color.New(color.FgYellow).SprintFunc(),
	Error:   color.New(color.FgRed).SprintFunc(),
	Info:    color.New(color.FgCyan).SprintFunc(),
	Muted:   color.New(color.FgHiBlack).SprintFunc(),
	Bold:    color.New(color.Bold).SprintFunc(),
	Accent:  color.New(color.FgBlue).SprintFunc(),
	Brand:   color.New(color.FgCyan, color.Bold).SprintFunc(),
}

// T is the active theme, used by all output helpers.
var T = DefaultTheme

// Sym contains Unicode symbols for CLI output.
var Sym = struct {
	Success string
	Warning string
	Error   string
	Info    string
	Arrow   string
	Bullet  string
	BoxH    string
	Zap     string
}{
	Success: "\u2713", // ✓
	Warning: "\u26A0", // ⚠
	Error:   "\u2717", // ✗
	Info:    "\u2139", // ℹ
	Arrow:   "\u2192", // →
	Bullet:  "\u2022", // •
	BoxH:    "\u2500", // ─
	Zap:     "\u26A1", // ⚡
}

// StatusColor returns the given status string colorized by its meaning.
func StatusColor(status string) string {
	switch status {
	case "active":
		return T.Success(status)
	case "stale", "disconnected":
		return T.Warning(status)
	case "completed":
		return T.Muted(status)
	case "pending":
		return T.Info(status)
	case "failed":
		return T.Error(status)
	case "claimed", "in_progress":
		return T.Accent(status)
	default:
		return status
	}
}

// PriorityColor returns the given priority label colorized by urgency.
func PriorityColor(priority string) string {
	switch priority {
	case "critical":
		return T.Error(priority)
	case "high":
		return T.Warning(priority)
	default:
		return priority
	}
}

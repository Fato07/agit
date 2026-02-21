package ui

import (
	"os"

	"github.com/mattn/go-isatty"
)

// IsTerminal returns true when stdout is connected to a terminal.
func IsTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

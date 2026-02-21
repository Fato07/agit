package ui

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

// NewTable returns a pre-configured tablewriter with standard agit styling.
func NewTable(headers ...string) *tablewriter.Table {
	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader(headers)
	t.SetBorder(false)
	t.SetColumnSeparator("")
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	return t
}

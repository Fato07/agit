package conflicts

import (
	"fmt"
	"strings"

	"github.com/fathindos/agit/internal/registry"
)

// FormatReport returns a human-readable conflict report
func FormatReport(conflicts []registry.Conflict) string {
	if len(conflicts) == 0 {
		return "No conflicts detected."
	}

	var b strings.Builder
	for _, c := range conflicts {
		fmt.Fprintf(&b, "CONFLICT: %s\n", c.FilePath)
		for i, wtID := range c.Worktrees {
			shortID := wtID
			if len(shortID) > 12 {
				shortID = shortID[:12]
			}
			agent := ""
			if i < len(c.AgentIDs) && c.AgentIDs[i] != "" {
				agent = c.AgentIDs[i]
			}
			task := ""
			if i < len(c.TaskDescs) && c.TaskDescs[i] != "" {
				task = c.TaskDescs[i]
			}

			if agent != "" {
				fmt.Fprintf(&b, "  Modified in: %s (%s: %s)\n", shortID, agent, task)
			} else {
				fmt.Fprintf(&b, "  Modified in: %s\n", shortID)
			}
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "%d conflict(s) across %d file(s).\n", len(conflicts), len(conflicts))
	return b.String()
}

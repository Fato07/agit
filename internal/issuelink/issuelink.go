package issuelink

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"
)

const (
	RepoURL   = "https://github.com/fathindos/agit"
	MaxURLLen = 2000
	EnvOptOut = "AGIT_NO_ISSUE_LINK"
)

// AppVersion is set by cmd.Execute() at startup to avoid circular imports.
var AppVersion = "unknown"

// Context holds diagnostic information for building an issue URL.
type Context struct {
	Err     error
	Command []string
	Version string
}

// Enabled reports whether issue link generation is active.
func Enabled() bool {
	return os.Getenv(EnvOptOut) == ""
}

// ForError generates a pre-filled GitHub issue URL for an error,
// using the package-level AppVersion and os.Args for context.
func ForError(err error) string {
	return Build(Context{
		Err:     err,
		Command: os.Args,
		Version: AppVersion,
	})
}

// Build generates a pre-filled GitHub issue URL from the error context.
func Build(ctx Context) string {
	title := "Bug: " + firstLine(ctx.Err.Error())
	body := buildBody(ctx)
	return buildURL(title, body)
}

func buildBody(ctx Context) string {
	return fmt.Sprintf(
		"## Error\n```\n%s\n```\n\n## Environment\n- **agit version**: %s\n- **OS**: %s\n- **Arch**: %s\n- **Command**: `%s`\n",
		ctx.Err.Error(),
		ctx.Version,
		runtime.GOOS,
		runtime.GOARCH,
		strings.Join(ctx.Command, " "),
	)
}

func buildURL(title, body string) string {
	u, _ := url.Parse(RepoURL + "/issues/new")
	params := url.Values{}
	params.Set("title", title)
	params.Set("labels", "bug")
	params.Set("template", "bug_report.md")
	params.Set("body", body)

	u.RawQuery = params.Encode()
	full := u.String()
	if len(full) <= MaxURLLen {
		return full
	}
	return truncateURL(title, body)
}

func truncateURL(title, body string) string {
	note := "\n\n[body truncated — see terminal output for full error]"

	u, _ := url.Parse(RepoURL + "/issues/new")
	params := url.Values{}
	params.Set("title", title)
	params.Set("labels", "bug")
	params.Set("template", "bug_report.md")
	params.Set("body", note)

	u.RawQuery = params.Encode()
	baseLen := len(u.String())

	available := MaxURLLen - baseLen
	if available <= 0 {
		return u.String()
	}

	// Estimate how much raw body we can fit (URL encoding expands ~3x worst case)
	// Use a conservative approach: try progressively shorter bodies
	for tryLen := len(body); tryLen > 0; tryLen = tryLen * 3 / 4 {
		candidate := body[:tryLen] + note
		params.Set("body", candidate)
		u.RawQuery = params.Encode()
		if len(u.String()) <= MaxURLLen {
			return u.String()
		}
	}

	// Fallback: just the note
	params.Set("body", note)
	u.RawQuery = params.Encode()
	return u.String()
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

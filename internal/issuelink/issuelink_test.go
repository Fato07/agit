package issuelink_test

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/fathindos/agit/internal/issuelink"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name       string
		ctx        issuelink.Context
		wantTitle  string
		wantLabels string
		wantBody   string
	}{
		{
			name: "generates valid URL with title body and labels",
			ctx: issuelink.Context{
				Err:     errors.New("something broke"),
				Command: []string{"agit", "mcp"},
				Version: "1.0.0",
			},
			wantTitle:  "Bug: something broke",
			wantLabels: "bug",
			wantBody:   "something broke",
		},
		{
			name: "special characters are properly URL-encoded",
			ctx: issuelink.Context{
				Err:     errors.New("error: unexpected token '&' in <body>"),
				Command: []string{"agit", "run"},
				Version: "2.0.0",
			},
			wantTitle:  "Bug: error: unexpected token '&' in <body>",
			wantLabels: "bug",
			wantBody:   "error: unexpected token '&' in <body>",
		},
		{
			name: "multi-line error uses only first line in title",
			ctx: issuelink.Context{
				Err:     errors.New("first line error\nsecond line detail\nthird line"),
				Command: []string{"agit", "serve"},
				Version: "1.5.0",
			},
			wantTitle:  "Bug: first line error",
			wantLabels: "bug",
			wantBody:   "first line error\nsecond line detail\nthird line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := issuelink.Build(tt.ctx)

			u, err := url.Parse(result)
			if err != nil {
				t.Fatalf("Build() returned invalid URL: %v", err)
			}

			if !strings.HasPrefix(result, issuelink.RepoURL+"/issues/new") {
				t.Errorf("URL does not start with repo issues URL, got: %s", result)
			}

			params := u.Query()

			if got := params.Get("title"); got != tt.wantTitle {
				t.Errorf("title = %q, want %q", got, tt.wantTitle)
			}

			if got := params.Get("labels"); got != tt.wantLabels {
				t.Errorf("labels = %q, want %q", got, tt.wantLabels)
			}

			body := params.Get("body")
			if !strings.Contains(body, tt.wantBody) {
				t.Errorf("body does not contain %q, got: %s", tt.wantBody, body)
			}

			if got := params.Get("template"); got != "bug_report.md" {
				t.Errorf("template = %q, want %q", got, "bug_report.md")
			}
		})
	}
}

func TestBuild_Truncation(t *testing.T) {
	// Create an error message that is at least 3000 characters to force truncation.
	// Use a short first line so the title stays small; the long content goes in the body.
	longMsg := "short error\n" + strings.Repeat("x", 3000)
	ctx := issuelink.Context{
		Err:     errors.New(longMsg),
		Command: []string{"agit", "serve", "--long-flag"},
		Version: "1.0.0",
	}

	result := issuelink.Build(ctx)

	if len(result) > issuelink.MaxURLLen {
		t.Errorf("Build() URL length = %d, want <= %d", len(result), issuelink.MaxURLLen)
	}

	u, err := url.Parse(result)
	if err != nil {
		t.Fatalf("Build() returned invalid URL after truncation: %v", err)
	}

	body := u.Query().Get("body")
	if !strings.Contains(body, "truncated") {
		t.Errorf("truncated URL body should contain truncation note, got: %s", body)
	}
}

func TestEnabled(t *testing.T) {
	tests := []struct {
		name   string
		setEnv bool
		envVal string
		want   bool
	}{
		{
			name:   "returns true when env var is not set",
			setEnv: false,
			want:   true,
		},
		{
			name:   "returns false when env var is set to 1",
			setEnv: true,
			envVal: "1",
			want:   false,
		},
		{
			name:   "returns false when env var is set to empty-looking string",
			setEnv: true,
			envVal: "true",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(issuelink.EnvOptOut, tt.envVal)
			} else {
				// Ensure the env var is unset for this subtest.
				os.Unsetenv(issuelink.EnvOptOut)
			}

			if got := issuelink.Enabled(); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestForError(t *testing.T) {
	// Set a known AppVersion so the URL is deterministic.
	origVersion := issuelink.AppVersion
	issuelink.AppVersion = "test-version-1.2.3"
	t.Cleanup(func() { issuelink.AppVersion = origVersion })

	origArgs := os.Args
	os.Args = []string{"agit", "mcp", "--stdio"}
	t.Cleanup(func() { os.Args = origArgs })

	result := issuelink.ForError(errors.New("test failure"))

	u, err := url.Parse(result)
	if err != nil {
		t.Fatalf("ForError() returned invalid URL: %v", err)
	}

	params := u.Query()

	if got := params.Get("title"); got != "Bug: test failure" {
		t.Errorf("title = %q, want %q", got, "Bug: test failure")
	}

	body := params.Get("body")
	if !strings.Contains(body, "test-version-1.2.3") {
		t.Errorf("body should contain AppVersion, got: %s", body)
	}
	if !strings.Contains(body, "agit mcp --stdio") {
		t.Errorf("body should contain os.Args command, got: %s", body)
	}
}

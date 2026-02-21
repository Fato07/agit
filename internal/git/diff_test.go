package git

import (
	"testing"
)

func TestParseModifiedFiles(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name:   "normal output",
			output: "file1.go\nfile2.go\nfile3.go",
			want:   []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name:   "empty output",
			output: "",
			want:   nil,
		},
		{
			name:   "trailing newlines",
			output: "file1.go\nfile2.go\n\n",
			want:   []string{"file1.go", "file2.go"},
		},
		{
			name:   "single file",
			output: "main.go\n",
			want:   []string{"main.go"},
		},
		{
			name:   "whitespace only",
			output: "   \n  ",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseModifiedFiles(tt.output)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d files, want %d: %v", len(got), len(tt.want), got)
			}
			for i, f := range got {
				if f != tt.want[i] {
					t.Errorf("file[%d] = %q, want %q", i, f, tt.want[i])
				}
			}
		})
	}
}

func TestParseModifiedFilesWithStatus(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   map[string]string
	}{
		{
			name:   "mixed statuses",
			output: "A\tnew_file.go\nM\texisting.go\nD\tremoved.go\nR100\told.go",
			want: map[string]string{
				"new_file.go": "added",
				"existing.go": "modified",
				"removed.go":  "deleted",
				"old.go":      "renamed",
			},
		},
		{
			name:   "empty output",
			output: "",
			want:   map[string]string{},
		},
		{
			name:   "malformed lines ignored",
			output: "A\tfile.go\nbadline\nM\tother.go",
			want: map[string]string{
				"file.go":  "added",
				"other.go": "modified",
			},
		},
		{
			name:   "trailing newlines",
			output: "M\tfile.go\n\n",
			want: map[string]string{
				"file.go": "modified",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseModifiedFilesWithStatus(tt.output)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d files, want %d: %v", len(got), len(tt.want), got)
			}
			for path, wantStatus := range tt.want {
				if gotStatus, ok := got[path]; !ok {
					t.Errorf("missing file %q", path)
				} else if gotStatus != wantStatus {
					t.Errorf("file %q: got %q, want %q", path, gotStatus, wantStatus)
				}
			}
		})
	}
}

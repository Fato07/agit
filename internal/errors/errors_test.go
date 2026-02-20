package errors_test

import (
	"fmt"
	"testing"

	apperrors "github.com/fathindos/agit/internal/errors"
)

func TestNewUserError(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{
			name: "simple message",
			msg:  "invalid input",
			want: "invalid input",
		},
		{
			name: "empty message",
			msg:  "",
			want: "",
		},
		{
			name: "message with special characters",
			msg:  "field 'name' is required: missing value",
			want: "field 'name' is required: missing value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := apperrors.NewUserError(tt.msg)
			if err == nil {
				t.Fatal("NewUserError returned nil")
			}
			if got := err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewUserErrorf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
		want   string
	}{
		{
			name:   "formatted with string arg",
			format: "invalid field: %s",
			args:   []any{"name"},
			want:   "invalid field: name",
		},
		{
			name:   "formatted with int arg",
			format: "value %d out of range",
			args:   []any{42},
			want:   "value 42 out of range",
		},
		{
			name:   "formatted with multiple args",
			format: "%s must be between %d and %d",
			args:   []any{"age", 0, 120},
			want:   "age must be between 0 and 120",
		},
		{
			name:   "no format args",
			format: "plain message",
			args:   nil,
			want:   "plain message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := apperrors.NewUserErrorf(tt.format, tt.args...)
			if err == nil {
				t.Fatal("NewUserErrorf returned nil")
			}
			if got := err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsUserError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "UserError from NewUserError",
			err:  apperrors.NewUserError("bad input"),
			want: true,
		},
		{
			name: "UserError from NewUserErrorf",
			err:  apperrors.NewUserErrorf("bad %s", "input"),
			want: true,
		},
		{
			name: "regular error",
			err:  fmt.Errorf("some internal error"),
			want: false,
		},
		{
			name: "wrapped UserError",
			err:  fmt.Errorf("context: %w", apperrors.NewUserError("bad input")),
			want: true,
		},
		{
			name: "double wrapped UserError",
			err:  fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", apperrors.NewUserError("bad input"))),
			want: true,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apperrors.IsUserError(tt.err); got != tt.want {
				t.Errorf("IsUserError() = %v, want %v", got, tt.want)
			}
		})
	}
}

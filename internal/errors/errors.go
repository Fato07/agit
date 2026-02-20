package errors

import (
	"errors"
	"fmt"
)

// UserError marks an error as caused by invalid user input.
// These errors do not trigger issue link generation.
type UserError struct {
	msg string
}

func (e *UserError) Error() string { return e.msg }

// NewUserError creates a user input error.
func NewUserError(msg string) error {
	return &UserError{msg: msg}
}

// NewUserErrorf creates a formatted user input error.
func NewUserErrorf(format string, args ...any) error {
	return &UserError{msg: fmt.Sprintf(format, args...)}
}

// IsUserError reports whether any error in the chain is a UserError.
func IsUserError(err error) bool {
	var u *UserError
	return errors.As(err, &u)
}

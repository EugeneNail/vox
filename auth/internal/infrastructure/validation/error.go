package validation

import "fmt"

// Error aggregates field validation violations.
type Error struct {
	violations map[string]string
}

// NewError creates an empty validation error container.
func NewError() Error {
	return Error{
		violations: map[string]string{},
	}
}

// Error implements the error interface.
func (error Error) Error() string {
	return fmt.Sprintf("validation failed: %d violations", len(error.violations))
}

// Violations returns collected field error messages.
func (error Error) Violations() map[string]string {
	return error.violations
}

// AddViolation stores a validation message for the field.
func (error Error) AddViolation(field string, message string) {
	error.violations[field] = message
}

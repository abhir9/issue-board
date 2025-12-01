package validator

import (
	"fmt"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var messages []string
	for _, err := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// Validator provides validation functions
type Validator struct {
	errors ValidationErrors
}

// New creates a new Validator
func New() *Validator {
	return &Validator{
		errors: ValidationErrors{},
	}
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Required checks if a string is not empty
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
	}
}

// MaxLength checks if a string does not exceed max length
func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.AddError(field, fmt.Sprintf("must not exceed %d characters", max))
	}
}

// MinLength checks if a string meets minimum length
func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.AddError(field, fmt.Sprintf("must be at least %d characters", min))
	}
}

// OneOf checks if value is one of allowed values
func (v *Validator) OneOf(field, value string, allowed []string) {
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.AddError(field, fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
}

// Valid returns true if there are no validation errors
func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

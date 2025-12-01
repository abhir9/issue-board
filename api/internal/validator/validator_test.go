package validator

import (
	"strings"
	"testing"
)

func TestRequired(t *testing.T) {
	t.Run("Valid non-empty string", func(t *testing.T) {
		v := New()
		v.Required("field", "value")

		if !v.Valid() {
			t.Error("Expected valid for non-empty string")
		}
	})

	t.Run("Empty string", func(t *testing.T) {
		v := New()
		v.Required("field", "")

		if v.Valid() {
			t.Error("Expected invalid for empty string")
		}

		errors := v.Errors()
		if len(errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(errors))
		}

		if errors[0].Field != "field" {
			t.Errorf("Expected field 'field', got '%s'", errors[0].Field)
		}
	})

	t.Run("Whitespace only string", func(t *testing.T) {
		v := New()
		v.Required("field", "   ")

		if v.Valid() {
			t.Error("Expected invalid for whitespace-only string")
		}
	})
}

func TestMaxLength(t *testing.T) {
	t.Run("Within max length", func(t *testing.T) {
		v := New()
		v.MaxLength("field", "hello", 10)

		if !v.Valid() {
			t.Error("Expected valid for string within max length")
		}
	})

	t.Run("Exceeds max length", func(t *testing.T) {
		v := New()
		v.MaxLength("field", "this is a very long string", 10)

		if v.Valid() {
			t.Error("Expected invalid for string exceeding max length")
		}

		errors := v.Errors()
		if len(errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(errors))
		}

		if !strings.Contains(errors[0].Message, "10") {
			t.Error("Expected error message to contain max length")
		}
	})

	t.Run("Exactly max length", func(t *testing.T) {
		v := New()
		v.MaxLength("field", "exactly10c", 10)

		if !v.Valid() {
			t.Error("Expected valid for string exactly at max length")
		}
	})
}

func TestMinLength(t *testing.T) {
	t.Run("Meets min length", func(t *testing.T) {
		v := New()
		v.MinLength("field", "hello", 3)

		if !v.Valid() {
			t.Error("Expected valid for string meeting min length")
		}
	})

	t.Run("Below min length", func(t *testing.T) {
		v := New()
		v.MinLength("field", "hi", 5)

		if v.Valid() {
			t.Error("Expected invalid for string below min length")
		}

		errors := v.Errors()
		if len(errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(errors))
		}

		if !strings.Contains(errors[0].Message, "5") {
			t.Error("Expected error message to contain min length")
		}
	})

	t.Run("Exactly min length", func(t *testing.T) {
		v := New()
		v.MinLength("field", "12345", 5)

		if !v.Valid() {
			t.Error("Expected valid for string exactly at min length")
		}
	})
}

func TestOneOf(t *testing.T) {
	allowed := []string{"Low", "Medium", "High", "Critical"}

	t.Run("Valid value", func(t *testing.T) {
		v := New()
		v.OneOf("priority", "High", allowed)

		if !v.Valid() {
			t.Error("Expected valid for value in allowed list")
		}
	})

	t.Run("Invalid value", func(t *testing.T) {
		v := New()
		v.OneOf("priority", "Extreme", allowed)

		if v.Valid() {
			t.Error("Expected invalid for value not in allowed list")
		}

		errors := v.Errors()
		if len(errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(errors))
		}

		errorMsg := errors[0].Message
		for _, val := range allowed {
			if !strings.Contains(errorMsg, val) {
				t.Errorf("Expected error message to contain '%s'", val)
			}
		}
	})

	t.Run("Case sensitive", func(t *testing.T) {
		v := New()
		v.OneOf("priority", "high", allowed)

		if v.Valid() {
			t.Error("Expected invalid for case-mismatched value")
		}
	})
}

func TestMultipleValidations(t *testing.T) {
	t.Run("All validations pass", func(t *testing.T) {
		v := New()
		v.Required("title", "My Title")
		v.MinLength("title", "My Title", 3)
		v.MaxLength("title", "My Title", 50)

		if !v.Valid() {
			t.Error("Expected valid when all validations pass")
		}

		if len(v.Errors()) != 0 {
			t.Errorf("Expected 0 errors, got %d", len(v.Errors()))
		}
	})

	t.Run("Multiple validations fail", func(t *testing.T) {
		v := New()
		v.Required("field1", "")
		v.MaxLength("field2", "this is way too long for the limit", 10)
		v.OneOf("field3", "invalid", []string{"valid1", "valid2"})

		if v.Valid() {
			t.Error("Expected invalid when multiple validations fail")
		}

		errors := v.Errors()
		if len(errors) != 3 {
			t.Errorf("Expected 3 errors, got %d", len(errors))
		}
	})
}

func TestAddError(t *testing.T) {
	v := New()

	v.AddError("custom_field", "custom error message")

	if v.Valid() {
		t.Error("Expected invalid after adding custom error")
	}

	errors := v.Errors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if errors[0].Field != "custom_field" {
		t.Errorf("Expected field 'custom_field', got '%s'", errors[0].Field)
	}

	if errors[0].Message != "custom error message" {
		t.Errorf("Expected message 'custom error message', got '%s'", errors[0].Message)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errors := ValidationErrors{
		{Field: "field1", Message: "is required"},
		{Field: "field2", Message: "must be at least 5 characters"},
	}

	errorStr := errors.Error()

	if !strings.Contains(errorStr, "field1") {
		t.Error("Expected error string to contain 'field1'")
	}

	if !strings.Contains(errorStr, "field2") {
		t.Error("Expected error string to contain 'field2'")
	}

	if !strings.Contains(errorStr, ";") {
		t.Error("Expected error string to contain separator ';'")
	}
}

func TestValidatorReset(t *testing.T) {
	v := New()
	v.Required("field", "")

	if v.Valid() {
		t.Error("Expected invalid after adding error")
	}

	// Create new validator (reset)
	v = New()

	if !v.Valid() {
		t.Error("Expected new validator to be valid")
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("Empty allowed list", func(t *testing.T) {
		v := New()
		v.OneOf("field", "value", []string{})

		if v.Valid() {
			t.Error("Expected invalid for empty allowed list")
		}
	})

	t.Run("Zero min length", func(t *testing.T) {
		v := New()
		v.MinLength("field", "", 0)

		if !v.Valid() {
			t.Error("Expected valid for zero min length")
		}
	})

	t.Run("Zero max length", func(t *testing.T) {
		v := New()
		v.MaxLength("field", "a", 0)

		if v.Valid() {
			t.Error("Expected invalid for non-empty string with zero max length")
		}
	})
}

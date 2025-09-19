package server

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
)

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
func timeStrPtr(t time.Time) *string {
	return stringPtr(t.Format(time.RFC3339))
}

func validationErrorMessage(err error) string {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return "Invalid request payload"
	}

	messages := make([]string, 0, len(validationErrs))
	for _, fieldErr := range validationErrs {
		field := lowerFirst(fieldErr.Field())
		var message string

		switch fieldErr.Tag() {
		case "required":
			message = fmt.Sprintf("%s is required", field)
		case "email":
			message = fmt.Sprintf("%s must be a valid email address", field)
		default:
			message = fmt.Sprintf("%s failed '%s' validation", field, fieldErr.Tag())
		}

		messages = append(messages, message)
	}

	return strings.Join(messages, ", ")
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

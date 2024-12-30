package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func CoordinateValidationErrors(validationErrors error) map[string][]string {
	errors := make(map[string][]string)

	for _, err := range validationErrors.(validator.ValidationErrors) {
		field := err.Field()

		errors[field] = []string{}

		switch err.ActualTag() {
		case "required":
			errors[field] = append(errors[field], fmt.Sprintf("%sは必須です", err.Field()))
		}
	}

	return errors
}

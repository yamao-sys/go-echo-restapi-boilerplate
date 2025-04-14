package validator

import (
	apis "app/openapi"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func ValidateCreateTodo(input apis.PostTodosJSONRequestBody) error {
	return validation.ValidateStruct(&input,
		validation.Field(
			&input.Title,
			validation.Required.Error("タイトルは必須入力です。"),
			validation.RuneLength(1, 50).Error("タイトルは1 ~ 50文字での入力をお願いします。"),
		),
	)
}

func ValidateUpdateTodo(input  apis.PatchTodoJSONRequestBody) error {
	return validation.ValidateStruct(&input,
		validation.Field(
			&input.Title,
			validation.Required.Error("タイトルは必須入力です。"),
			validation.RuneLength(1, 50).Error("タイトルは1 ~ 50文字での入力をお願いします。"),
		),
	)
}

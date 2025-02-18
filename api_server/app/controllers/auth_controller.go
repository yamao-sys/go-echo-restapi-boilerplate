package controllers

import (
	"app/generated/auth"
	"app/services"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4/middleware"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type AuthController interface {
	GetAuthCsrf(ctx context.Context, request auth.GetAuthCsrfRequestObject) (auth.GetAuthCsrfResponseObject, error)
	PostAuthSignIn(ctx context.Context, request auth.PostAuthSignInRequestObject) (auth.PostAuthSignInResponseObject, error)
	PostAuthValidateSignUp(ctx context.Context, request auth.PostAuthValidateSignUpRequestObject) (auth.PostAuthValidateSignUpResponseObject, error)
	PostAuthSignUp(ctx context.Context, request auth.PostAuthSignUpRequestObject) (auth.PostAuthSignUpResponseObject, error)
}

type authController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) AuthController {
	return &authController{authService}
}

func (authController *authController) GetAuthCsrf(ctx context.Context, request auth.GetAuthCsrfRequestObject) (auth.GetAuthCsrfResponseObject, error) {
	csrfToken, ok := ctx.Value(middleware.DefaultCSRFConfig.ContextKey).(string)
	if !ok {
		return auth.GetAuthCsrf500JSONResponse{InternalServerErrorResponseJSONResponse: auth.InternalServerErrorResponseJSONResponse{
			Code: http.StatusInternalServerError,
			Message: "failed to retrieval token",
		}}, nil
	}
	
	return auth.GetAuthCsrf200JSONResponse{CsrfResponseJSONResponse: auth.CsrfResponseJSONResponse{ CsrfToken: csrfToken }}, nil
}

func (authController *authController) PostAuthSignIn(ctx context.Context, request auth.PostAuthSignInRequestObject) (auth.PostAuthSignInResponseObject, error) {
	inputs := auth.PostAuthSignInJSONBody{
		Email: request.Body.Email,
		Password: request.Body.Password,
	}

	statusCode, tokenString, err := authController.authService.SignIn(ctx, inputs)
	switch (statusCode) {
	case http.StatusInternalServerError:
		return auth.PostAuthSignIn500JSONResponse{InternalServerErrorResponseJSONResponse: auth.InternalServerErrorResponseJSONResponse{
			Code: http.StatusInternalServerError,
			Message: err.Error(),
		}}, nil
	case http.StatusBadRequest:
		return auth.PostAuthSignIn400JSONResponse{SignInBadRequestResponseJSONResponse: auth.SignInBadRequestResponseJSONResponse{
			Errors: []string{err.Error()},
		}}, nil
	}
	
	// NOTE: Cookieにtokenをセット
	cookie := &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		MaxAge:   3600 * 24,
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: true,
	}
	return auth.PostAuthSignIn200JSONResponse{SignInOkResponseJSONResponse: auth.SignInOkResponseJSONResponse{
		Headers: auth.SignInOkResponseResponseHeaders{
			SetCookie: cookie.String(),
		},
	}}, nil
}

func (authController *authController) PostAuthValidateSignUp(ctx context.Context, request auth.PostAuthValidateSignUpRequestObject) (auth.PostAuthValidateSignUpResponseObject, error) {
	reader := request.Body
	// NOTE: バリデーションチェックを行う構造体
	inputStruct, mappingErr := authController.mappingInputStruct(reader)
	if mappingErr != nil {
		return auth.PostAuthValidateSignUp500JSONResponse{InternalServerErrorResponseJSONResponse: auth.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: mappingErr.Error()}}, nil
	}

	err := authController.authService.ValidateSignUp(ctx, &inputStruct)
	validationError := authController.mappingValidationErrorStruct(err)

	res := &auth.SignUpResponse{
		Code: http.StatusOK,
		Errors: validationError,
	}
	return auth.PostAuthValidateSignUp200JSONResponse{SignUpResponseJSONResponse: auth.SignUpResponseJSONResponse{Code: res.Code, Errors: res.Errors}}, nil
}

func (authController *authController) PostAuthSignUp(ctx context.Context, request auth.PostAuthSignUpRequestObject) (auth.PostAuthSignUpResponseObject, error) {
	reader := request.Body
	// NOTE: バリデーションチェックを行う構造体
	inputStruct, mappingErr := authController.mappingInputStruct(reader)
	if mappingErr != nil {
		return auth.PostAuthSignUp500JSONResponse{InternalServerErrorResponseJSONResponse: auth.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: mappingErr.Error()}}, nil
	}

	err := authController.authService.ValidateSignUp(ctx, &inputStruct)
	if err != nil {
		validationError := authController.mappingValidationErrorStruct(err)
	
		res := &auth.SignUpResponse{
			Code: http.StatusBadRequest,
			Errors: validationError,
		}
		return auth.PostAuthSignUp400JSONResponse{Code: res.Code, Errors: res.Errors}, nil
	}

	signUpErr := authController.authService.SignUp(ctx, auth.PostAuthSignUpMultipartRequestBody(inputStruct))
	if signUpErr != nil {
		log.Fatalln(err)
	}

	res := &auth.SignUpResponse{
		Code: http.StatusOK,
		Errors: auth.SignUpValidationError{},
	}
	return auth.PostAuthSignUp200JSONResponse{SignUpResponseJSONResponse: auth.SignUpResponseJSONResponse{Code: res.Code, Errors: res.Errors}}, nil
}

func (authController *authController) mappingInputStruct(reader *multipart.Reader) (auth.PostAuthValidateSignUpMultipartRequestBody, error) {
	var inputStruct auth.PostAuthValidateSignUpMultipartRequestBody

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			// NOTE: 全てのパートを読み終えた場合
			break
		}
		if err != nil {
			return inputStruct, fmt.Errorf("failed to read multipart part: %w", err)
		}

		// NOTE: 各パートのヘッダー情報を取得
		partName := part.FormName()
		filename := part.FileName()

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, part); err != nil {
			fmt.Println(err)
			return inputStruct, fmt.Errorf("failed to copy content: %w", err)
		}

		switch partName {
		case "firstName":
			inputStruct.FirstName = buf.String()
		case "lastName":
			inputStruct.LastName = buf.String()
		case "password":
			inputStruct.Password = buf.String()
		case "email":
			inputStruct.Email = buf.String()
		case "birthday":
			birthday := buf.String()
			if birthday == "" {
				continue
			}
			
			parsedTime, parseErr := time.Parse("2006-01-02", birthday)
			if parseErr != nil {
				fmt.Println(parseErr)
				return inputStruct, fmt.Errorf("failed to parse birthday: %w", parseErr)
			}
			inputStruct.Birthday = &openapi_types.Date{Time: parsedTime}
		case "frontIdentification":
			var frontIdentification openapi_types.File
			frontIdentification.InitFromBytes(buf.Bytes(), filename)
			inputStruct.FrontIdentification = &frontIdentification
		case "backIdentification":
			var backIdentification openapi_types.File
			backIdentification.InitFromBytes(buf.Bytes(), filename)
			inputStruct.BackIdentification = &backIdentification
		}
	}

	return inputStruct, nil
}

func (authController *authController) mappingValidationErrorStruct(err error) auth.SignUpValidationError {
	var validationError auth.SignUpValidationError
	if err == nil {
		return validationError
	}

	if errors, ok := err.(validation.Errors); ok {
		// NOTE: レスポンス用の構造体にマッピング
		for field, err := range errors {
			messages := []string{err.Error()}
			switch field {
			case "firstName":
				validationError.FirstName = &messages
			case "lastName":
				validationError.LastName = &messages
			case "email":
				validationError.Email = &messages
			case "password":
				validationError.Password = &messages
			case "frontIdentification":
				validationError.FrontIdentification = &messages
			case "backIdentification":
				validationError.BackIdentification = &messages
			}
		}
	}
	return validationError
}

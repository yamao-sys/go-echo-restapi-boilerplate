package handlers

import (
	apis "app/openapi"
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

type AuthHandler interface {
	GetAuthCsrf(ctx context.Context, request apis.GetAuthCsrfRequestObject) (apis.GetAuthCsrfResponseObject, error)
	PostAuthSignIn(ctx context.Context, request apis.PostAuthSignInRequestObject) (apis.PostAuthSignInResponseObject, error)
	PostAuthValidateSignUp(ctx context.Context, request apis.PostAuthValidateSignUpRequestObject) (apis.PostAuthValidateSignUpResponseObject, error)
	PostAuthSignUp(ctx context.Context, request apis.PostAuthSignUpRequestObject) (apis.PostAuthSignUpResponseObject, error)
}

type authHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) AuthHandler {
	return &authHandler{authService}
}

func (authHandler *authHandler) GetAuthCsrf(ctx context.Context, request apis.GetAuthCsrfRequestObject) (apis.GetAuthCsrfResponseObject, error) {
	csrfToken, ok := ctx.Value(middleware.DefaultCSRFConfig.ContextKey).(string)
	if !ok {
		return apis.GetAuthCsrf500JSONResponse{InternalServerErrorResponseJSONResponse: apis.InternalServerErrorResponseJSONResponse{
			Code: http.StatusInternalServerError,
			Message: "failed to retrieval token",
		}}, nil
	}
	
	return apis.GetAuthCsrf200JSONResponse{CsrfResponseJSONResponse: apis.CsrfResponseJSONResponse{ CsrfToken: csrfToken }}, nil
}

func (authHandler *authHandler) PostAuthSignIn(ctx context.Context, request apis.PostAuthSignInRequestObject) (apis.PostAuthSignInResponseObject, error) {
	inputs := apis.PostAuthSignInJSONBody{
		Email: request.Body.Email,
		Password: request.Body.Password,
	}

	statusCode, tokenString, err := authHandler.authService.SignIn(ctx, inputs)
	switch (statusCode) {
	case http.StatusInternalServerError:
		return apis.PostAuthSignIn500JSONResponse{InternalServerErrorResponseJSONResponse: apis.InternalServerErrorResponseJSONResponse{
			Code: http.StatusInternalServerError,
			Message: err.Error(),
		}}, nil
	case http.StatusBadRequest:
		return apis.PostAuthSignIn400JSONResponse{SignInBadRequestResponseJSONResponse: apis.SignInBadRequestResponseJSONResponse{
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
	return apis.PostAuthSignIn200JSONResponse{SignInOkResponseJSONResponse: apis.SignInOkResponseJSONResponse{
		Headers: apis.SignInOkResponseResponseHeaders{
			SetCookie: cookie.String(),
		},
	}}, nil
}

func (authHandler *authHandler) PostAuthValidateSignUp(ctx context.Context, request apis.PostAuthValidateSignUpRequestObject) (apis.PostAuthValidateSignUpResponseObject, error) {
	reader := request.Body
	// NOTE: バリデーションチェックを行う構造体
	inputStruct, mappingErr := authHandler.mappingInputStruct(reader)
	if mappingErr != nil {
		return apis.PostAuthValidateSignUp500JSONResponse{InternalServerErrorResponseJSONResponse: apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: mappingErr.Error()}}, nil
	}

	err := authHandler.authService.ValidateSignUp(ctx, &inputStruct)
	validationError := authHandler.mappingValidationErrorStruct(err)

	res := &apis.SignUpResponse{
		Code: http.StatusOK,
		Errors: validationError,
	}
	return apis.PostAuthValidateSignUp200JSONResponse{SignUpResponseJSONResponse: apis.SignUpResponseJSONResponse{Code: res.Code, Errors: res.Errors}}, nil
}

func (authHandler *authHandler) PostAuthSignUp(ctx context.Context, request apis.PostAuthSignUpRequestObject) (apis.PostAuthSignUpResponseObject, error) {
	reader := request.Body
	// NOTE: バリデーションチェックを行う構造体
	inputStruct, mappingErr := authHandler.mappingInputStruct(reader)
	if mappingErr != nil {
		return apis.PostAuthSignUp500JSONResponse{InternalServerErrorResponseJSONResponse: apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: mappingErr.Error()}}, nil
	}

	err := authHandler.authService.ValidateSignUp(ctx, &inputStruct)
	if err != nil {
		validationError := authHandler.mappingValidationErrorStruct(err)
	
		res := &apis.SignUpResponse{
			Code: http.StatusBadRequest,
			Errors: validationError,
		}
		return apis.PostAuthSignUp400JSONResponse{Code: res.Code, Errors: res.Errors}, nil
	}

	signUpErr := authHandler.authService.SignUp(ctx, apis.PostAuthSignUpMultipartRequestBody(inputStruct))
	if signUpErr != nil {
		log.Fatalln(err)
	}

	res := &apis.SignUpResponse{
		Code: http.StatusOK,
		Errors: apis.SignUpValidationError{},
	}
	return apis.PostAuthSignUp200JSONResponse{SignUpResponseJSONResponse: apis.SignUpResponseJSONResponse{Code: res.Code, Errors: res.Errors}}, nil
}

func (authHandler *authHandler) mappingInputStruct(reader *multipart.Reader) (apis.PostAuthValidateSignUpMultipartRequestBody, error) {
	var inputStruct apis.PostAuthValidateSignUpMultipartRequestBody

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

func (authHandler *authHandler) mappingValidationErrorStruct(err error) apis.SignUpValidationError {
	var validationError apis.SignUpValidationError
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

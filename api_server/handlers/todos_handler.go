package handlers

import (
	apis "app/openapi"
	"app/services"
	"app/utils"
	"context"
	"errors"
	"net/http"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type TodosHandler interface {
	GetTodos(ctx context.Context, request apis.GetTodosRequestObject) (apis.GetTodosResponseObject, error)
	PostTodos(ctx context.Context, request apis.PostTodosRequestObject) (apis.PostTodosResponseObject, error)
	GetTodo(ctx context.Context, request apis.GetTodoRequestObject) (apis.GetTodoResponseObject, error)
	PatchTodo(ctx context.Context, request apis.PatchTodoRequestObject) (apis.PatchTodoResponseObject, error)
	DeleteTodo(ctx context.Context, request apis.DeleteTodoRequestObject) (apis.DeleteTodoResponseObject, error)
}

type todosHandler struct {
	todoService services.TodoService
}

func NewTodosHandler(todoService services.TodoService) TodosHandler {
	return &todosHandler{todoService: todoService}
}

func (todosHandler *todosHandler) GetTodos(ctx context.Context, request apis.GetTodosRequestObject) (apis.GetTodosResponseObject, error) {
	userID, ok := utils.ContextValue(ctx)
	if !ok {
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return apis.GetTodos500JSONResponse{InternalServerErrorResponseJSONResponse: res}, errors.New("fail to load context value")
	}

	statusCode, todosList, err := todosHandler.todoService.FetchTodosList(ctx, userID)
	switch statusCode {
	case http.StatusInternalServerError:
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: err.Error()}
		return apis.GetTodos500JSONResponse{InternalServerErrorResponseJSONResponse: res}, err
	}

	var resTodosList apis.FetchTodosResponseJSONResponse
	for _, todo := range *todosList {
		resTodosList.Todos = append(resTodosList.Todos, apis.Todo{Id: int(todo.ID), Title: todo.Title, Content: todo.Content.String })
	}
	return apis.GetTodos200JSONResponse{FetchTodosResponseJSONResponse: resTodosList}, nil
}

func (todosHandler *todosHandler) PostTodos(ctx context.Context, request apis.PostTodosRequestObject) (apis.PostTodosResponseObject, error) {
	userID, ok := utils.ContextValue(ctx)
	if !ok {
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return apis.PostTodos500JSONResponse{InternalServerErrorResponseJSONResponse: res}, errors.New("fail to load context value")
	}
	statusCode, err := todosHandler.todoService.CreateTodo(ctx, *request.Body, userID)

	switch statusCode {
	case http.StatusBadRequest:
		validationErrors := todosHandler.mappingValidationErrorStruct(err)
		return apis.PostTodos400JSONResponse{Code: http.StatusBadRequest, Errors: validationErrors}, nil
	case http.StatusInternalServerError:
		return apis.PostTodos400JSONResponse{Code: http.StatusInternalServerError, Errors: apis.StoreTodoValidationError{}}, err
	}

	res := apis.StoreTodoResponseJSONResponse{ Code: http.StatusOK, Errors: apis.StoreTodoValidationError{} }
	return apis.PostTodos200JSONResponse{StoreTodoResponseJSONResponse: res}, nil
}

func (todosHandler *todosHandler) GetTodo(ctx context.Context, request apis.GetTodoRequestObject) (apis.GetTodoResponseObject, error) {
	intID, err := strconv.Atoi(request.Id)
	if err != nil {
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: err.Error()}
		return apis.GetTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, err
	}

	userID, ok := utils.ContextValue(ctx)
	if !ok {
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return apis.GetTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, errors.New("fail to load context value")
	}

	statusCode, todo := todosHandler.todoService.ShowTodo(ctx, int64(intID), userID)
	switch statusCode {
	case http.StatusNotFound:
		res := apis.NotFoundErrorResponseJSONResponse{Code: http.StatusNotFound}
		return apis.GetTodo404JSONResponse{NotFoundErrorResponseJSONResponse: res}, nil
	}

	resTodo := apis.Todo{Id: int(todo.ID), Title: todo.Title, Content: todo.Content.String}
	res := apis.ShowTodoResponseJSONResponse{Todo: resTodo}
	return apis.GetTodo200JSONResponse{ShowTodoResponseJSONResponse: res}, nil
}

func (todosHandler *todosHandler) PatchTodo(ctx context.Context, request apis.PatchTodoRequestObject) (apis.PatchTodoResponseObject, error) {
	intID, err := strconv.Atoi(request.Id)
	if err != nil {
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: err.Error()}
		return apis.PatchTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, err
	}

	userID, ok := utils.ContextValue(ctx)
	if !ok {
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return apis.PatchTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, errors.New("fail to load context value")
	}

	statusCode, err := todosHandler.todoService.UpdateTodo(ctx, int64(intID), *request.Body, userID)

	switch statusCode {
	case http.StatusBadRequest:
		validationErrors := todosHandler.mappingValidationErrorStruct(err)
		return apis.PatchTodo400JSONResponse{Code: http.StatusBadRequest, Errors: validationErrors}, nil
	case http.StatusNotFound:
		res := apis.NotFoundErrorResponseJSONResponse{Code: http.StatusNotFound}
		return apis.PatchTodo404JSONResponse{NotFoundErrorResponseJSONResponse: res}, nil
	case http.StatusInternalServerError:
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return apis.PatchTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, err
	}

	res := apis.StoreTodoResponseJSONResponse{ Code: http.StatusOK, Errors: apis.StoreTodoValidationError{} }
	return apis.PatchTodo200JSONResponse{StoreTodoResponseJSONResponse: res}, nil
}

func (todosHandler *todosHandler) DeleteTodo(ctx context.Context, request apis.DeleteTodoRequestObject) (apis.DeleteTodoResponseObject, error) {
	intID, err := strconv.Atoi(request.Id)
	if err != nil {
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: err.Error()}
		return apis.DeleteTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, err
	}

	userID, ok := utils.ContextValue(ctx)
	if !ok {
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return apis.DeleteTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, errors.New("fail to load context value")
	}

	statusCode, err := todosHandler.todoService.DeleteTodo(ctx, int64(intID), userID)

	switch statusCode {
	case http.StatusNotFound:
		res := apis.NotFoundErrorResponseJSONResponse{Code: http.StatusNotFound}
		return apis.DeleteTodo404JSONResponse{NotFoundErrorResponseJSONResponse: res}, nil
	case http.StatusInternalServerError:
		res := apis.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return apis.DeleteTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, err
	}

	res := apis.DeleteTodoResponseJSONResponse{ Code: http.StatusOK, Result: true }
	return apis.DeleteTodo200JSONResponse{DeleteTodoResponseJSONResponse: res}, nil
}

func (todosHandler *todosHandler) mappingValidationErrorStruct(err error) apis.StoreTodoValidationError {
	var validationError apis.StoreTodoValidationError
	if err == nil {
		return validationError
	}

	if errors, ok := err.(validation.Errors); ok {
		// NOTE: レスポンス用の構造体にマッピング
		for field, err := range errors {
			messages := []string{err.Error()}
			switch field {
			case "title":
				validationError.Title = &messages
			case "content":
				validationError.Content = &messages
			}
		}
	}
	return validationError
}

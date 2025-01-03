package controllers

import (
	"app/generated/todos"
	"app/services"
	"app/utils"
	"context"
	"errors"
	"net/http"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type TodosController interface {
	GetTodos(ctx context.Context, request todos.GetTodosRequestObject) (todos.GetTodosResponseObject, error)
	PostTodos(ctx context.Context, request todos.PostTodosRequestObject) (todos.PostTodosResponseObject, error)
	GetTodo(ctx context.Context, request todos.GetTodoRequestObject) (todos.GetTodoResponseObject, error)
	PatchTodo(ctx context.Context, request todos.PatchTodoRequestObject) (todos.PatchTodoResponseObject, error)
}

type todosController struct {
	todoService services.TodoService
}

func NewTodosController(todoService services.TodoService) TodosController {
	return &todosController{todoService: todoService}
}

func (todosController *todosController) GetTodos(ctx context.Context, request todos.GetTodosRequestObject) (todos.GetTodosResponseObject, error) {
	userID, ok := utils.ContextValue(ctx)
	if !ok {
		res := todos.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return todos.GetTodos500JSONResponse{InternalServerErrorResponseJSONResponse: res}, errors.New("fail to load context value")
	}

	statusCode, todosList, err := todosController.todoService.FetchTodosList(ctx, userID)
	switch statusCode {
	case http.StatusInternalServerError:
		res := todos.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: err.Error()}
		return todos.GetTodos500JSONResponse{InternalServerErrorResponseJSONResponse: res}, err
	}

	var resTodosList todos.FetchTodosResponseJSONResponse
	for _, todo := range *todosList {
		resTodosList.Todos = append(resTodosList.Todos, todos.Todo{Id: int(todo.ID), Title: todo.Title, Content: todo.Content.String })
	}
	return todos.GetTodos200JSONResponse{FetchTodosResponseJSONResponse: resTodosList}, nil
}

func (todosController *todosController) PostTodos(ctx context.Context, request todos.PostTodosRequestObject) (todos.PostTodosResponseObject, error) {
	userID, ok := utils.ContextValue(ctx)
	if !ok {
		res := todos.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return todos.PostTodos500JSONResponse{InternalServerErrorResponseJSONResponse: res}, errors.New("fail to load context value")
	}
	statusCode, err := todosController.todoService.CreateTodo(ctx, *request.Body, userID)

	switch statusCode {
	case http.StatusBadRequest:
		validationErrors := todosController.mappingValidationErrorStruct(err)
		return todos.PostTodos400JSONResponse{Code: http.StatusBadRequest, Errors: validationErrors}, nil
	case http.StatusInternalServerError:
		return todos.PostTodos400JSONResponse{Code: http.StatusInternalServerError, Errors: todos.StoreTodoValidationError{}}, err
	}

	res := todos.StoreTodoResponseJSONResponse{ Code: http.StatusOK, Errors: todos.StoreTodoValidationError{} }
	return todos.PostTodos200JSONResponse{StoreTodoResponseJSONResponse: res}, nil
}

func (todosController *todosController) GetTodo(ctx context.Context, request todos.GetTodoRequestObject) (todos.GetTodoResponseObject, error) {
	intID, err := strconv.Atoi(request.Id)
	if err != nil {
		res := todos.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: err.Error()}
		return todos.GetTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, err
	}

	userID, ok := utils.ContextValue(ctx)
	if !ok {
		res := todos.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return todos.GetTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, errors.New("fail to load context value")
	}

	statusCode, todo := todosController.todoService.ShowTodo(ctx, int64(intID), userID)
	switch statusCode {
	case http.StatusNotFound:
		res := todos.NotFoundErrorResponseJSONResponse{Code: http.StatusNotFound}
		return todos.GetTodo404JSONResponse{NotFoundErrorResponseJSONResponse: res}, nil
	}

	resTodo := todos.Todo{Id: int(todo.ID), Title: todo.Title, Content: todo.Content.String}
	res := todos.ShowTodoResponseJSONResponse{Todo: resTodo}
	return todos.GetTodo200JSONResponse{ShowTodoResponseJSONResponse: res}, nil
}

func (todosController *todosController) PatchTodo(ctx context.Context, request todos.PatchTodoRequestObject) (todos.PatchTodoResponseObject, error) {
	intID, err := strconv.Atoi(request.Id)
	if err != nil {
		res := todos.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError, Message: err.Error()}
		return todos.PatchTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, err
	}

	userID, ok := utils.ContextValue(ctx)
	if !ok {
		res := todos.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return todos.PatchTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, errors.New("fail to load context value")
	}

	statusCode, err := todosController.todoService.UpdateTodo(ctx, int64(intID), *request.Body, userID)

	switch statusCode {
	case http.StatusBadRequest:
		validationErrors := todosController.mappingValidationErrorStruct(err)
		return todos.PatchTodo400JSONResponse{Code: http.StatusBadRequest, Errors: validationErrors}, nil
	case http.StatusNotFound:
		res := todos.NotFoundErrorResponseJSONResponse{Code: http.StatusNotFound}
		return todos.PatchTodo404JSONResponse{NotFoundErrorResponseJSONResponse: res}, nil
	case http.StatusInternalServerError:
		res := todos.InternalServerErrorResponseJSONResponse{Code: http.StatusInternalServerError}
		return todos.PatchTodo500JSONResponse{InternalServerErrorResponseJSONResponse: res}, err
	}

	res := todos.StoreTodoResponseJSONResponse{ Code: http.StatusOK, Errors: todos.StoreTodoValidationError{} }
	return todos.PatchTodo200JSONResponse{StoreTodoResponseJSONResponse: res}, nil
}

func (todosController *todosController) mappingValidationErrorStruct(err error) todos.StoreTodoValidationError {
	var validationError todos.StoreTodoValidationError
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

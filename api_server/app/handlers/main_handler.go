package handlers

import (
	apis "app/openapi"
	"context"
)

type MainHandler interface {
	// handlers /auth
	GetAuthCsrf(ctx context.Context, request apis.GetAuthCsrfRequestObject) (apis.GetAuthCsrfResponseObject, error)
	PostAuthSignIn(ctx context.Context, request apis.PostAuthSignInRequestObject) (apis.PostAuthSignInResponseObject, error)
	PostAuthValidateSignUp(ctx context.Context, request apis.PostAuthValidateSignUpRequestObject) (apis.PostAuthValidateSignUpResponseObject, error)
	PostAuthSignUp(ctx context.Context, request apis.PostAuthSignUpRequestObject) (apis.PostAuthSignUpResponseObject, error)

	// handlers /todos
	GetTodos(ctx context.Context, request apis.GetTodosRequestObject) (apis.GetTodosResponseObject, error)
	PostTodos(ctx context.Context, request apis.PostTodosRequestObject) (apis.PostTodosResponseObject, error)
	GetTodo(ctx context.Context, request apis.GetTodoRequestObject) (apis.GetTodoResponseObject, error)
	PatchTodo(ctx context.Context, request apis.PatchTodoRequestObject) (apis.PatchTodoResponseObject, error)
	DeleteTodo(ctx context.Context, request apis.DeleteTodoRequestObject) (apis.DeleteTodoResponseObject, error)
}

type mainHandler struct {
	authHandler AuthHandler
	todosHandler TodosHandler
}

func NewMainHandler(authHandler AuthHandler, todosHandler TodosHandler) MainHandler {
	return &mainHandler{authHandler: authHandler, todosHandler: todosHandler}
}

func (mh *mainHandler) GetAuthCsrf(ctx context.Context, request apis.GetAuthCsrfRequestObject) (apis.GetAuthCsrfResponseObject, error) {
	res, err := mh.authHandler.GetAuthCsrf(ctx, request)
	return res, err
}

func (mh *mainHandler) PostAuthSignIn(ctx context.Context, request apis.PostAuthSignInRequestObject) (apis.PostAuthSignInResponseObject, error) {
	res, err := mh.authHandler.PostAuthSignIn(ctx, request)
	return res, err
}

func (mh *mainHandler) PostAuthValidateSignUp(ctx context.Context, request apis.PostAuthValidateSignUpRequestObject) (apis.PostAuthValidateSignUpResponseObject, error) {
	res, err := mh.authHandler.PostAuthValidateSignUp(ctx, request)
	return res, err
}

func (mh *mainHandler) PostAuthSignUp(ctx context.Context, request apis.PostAuthSignUpRequestObject) (apis.PostAuthSignUpResponseObject, error) {
	res, err := mh.authHandler.PostAuthSignUp(ctx, request)
	return res, err
}

func (mh *mainHandler) GetTodos(ctx context.Context, request apis.GetTodosRequestObject) (apis.GetTodosResponseObject, error) {
	res, err := mh.todosHandler.GetTodos(ctx, request)
	return res, err
}

func (mh *mainHandler) PostTodos(ctx context.Context, request apis.PostTodosRequestObject) (apis.PostTodosResponseObject, error) {
	res, err := mh.todosHandler.PostTodos(ctx, request)
	return res, err
}

func (mh *mainHandler) GetTodo(ctx context.Context, request apis.GetTodoRequestObject) (apis.GetTodoResponseObject, error) {
	res, err := mh.todosHandler.GetTodo(ctx, request)
	return res, err
}

func (mh *mainHandler) PatchTodo(ctx context.Context, request apis.PatchTodoRequestObject) (apis.PatchTodoResponseObject, error) {
	res, err := mh.todosHandler.PatchTodo(ctx, request)
	return res, err
}

func (mh *mainHandler) DeleteTodo(ctx context.Context, request apis.DeleteTodoRequestObject) (apis.DeleteTodoResponseObject, error) {
	res, err := mh.todosHandler.DeleteTodo(ctx, request)
	return res, err
}

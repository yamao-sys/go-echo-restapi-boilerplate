package services

import (
	models "app/models/generated"
	apis "app/openapi"
	"app/validator"
	"context"
	"database/sql"
	"net/http"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type TodoService interface {
	CreateTodo(ctx context.Context, requestParams apis.PostTodosJSONRequestBody, userID int64) (statusCode int, err error)
	FetchTodosList(ctx context.Context, userID int64) (statusCode int, todosList *models.TodoSlice, err error)
	ShowTodo(ctx context.Context, id int64, userID int64) (statusCode int, todo *models.Todo)
	UpdateTodo(ctx context.Context, id int64, requestParams apis.PatchTodoJSONRequestBody, userID int64) (statusCode int, err error)
	DeleteTodo(ctx context.Context, id int64, userID int64) (statusCode int, err error)
}

type todoService struct {
	db *sql.DB
}

func NewTodoService(db *sql.DB) TodoService {
	return &todoService{db}
}

func (ts *todoService) CreateTodo(ctx context.Context, requestParams apis.PostTodosJSONRequestBody, userID int64) (statusCode int, err error) {
	// NOTE: バリデーションチェック
	validationErrors := validator.ValidateCreateTodo(requestParams)
	if validationErrors != nil {
		return int(http.StatusBadRequest), validationErrors
	}

	todo := &models.Todo{}
	todo.Title = requestParams.Title
	todo.Content = null.String{String: requestParams.Content, Valid: true}
	todo.UserID = userID
	// NOTE: Create処理
	err = todo.Insert(ctx, ts.db, boil.Infer())
	if err != nil {
		return int(http.StatusInternalServerError), err
	}
	return int(http.StatusOK), nil
}

func (ts *todoService) FetchTodosList(ctx context.Context, userID int64) (statusCode int, todosList *models.TodoSlice, err error) {
	todos, err := models.Todos(qm.Where("user_id = ?", userID)).All(ctx, ts.db)
	if err != nil {
		return int(http.StatusInternalServerError), &models.TodoSlice{}, err
	}

	return int(http.StatusOK), &todos, nil
}

func (ts *todoService) ShowTodo(ctx context.Context, id int64, userID int64) (statusCode int, todo *models.Todo) {
	todo, err := models.Todos(qm.Where("id = ? AND user_id = ?", id, userID)).One(ctx, ts.db)
	if err != nil {
		return http.StatusNotFound, &models.Todo{}
	}

	return http.StatusOK, todo
}

func (ts *todoService) UpdateTodo(ctx context.Context, id int64, requestParams apis.PatchTodoJSONRequestBody, userID int64) (statusCode int, err error) {
	todo, err := models.Todos(qm.Where("id = ? AND user_id = ?", id, userID)).One(ctx, ts.db)
	if err != nil {
		return http.StatusNotFound, err
	}

	// NOTE: バリデーションチェック
	validationErrors := validator.ValidateUpdateTodo(requestParams)
	if validationErrors != nil {
		return int(http.StatusBadRequest), validationErrors
	}

	todo.Title = requestParams.Title
	todo.Content = null.String{String: requestParams.Content, Valid: true}

	// NOTE: Update処理
	_, updateError := todo.Update(ctx, ts.db, boil.Infer())
	if updateError != nil {
		return http.StatusInternalServerError, updateError
	}
	return http.StatusOK, nil
}

func (ts *todoService) DeleteTodo(ctx context.Context, id int64, userID int64) (statusCode int, err error) {
	todo, err := models.Todos(qm.Where("id = ? AND user_id = ?", id, userID)).One(ctx, ts.db)
	if err != nil {
		return http.StatusNotFound, err
	}

	_, deleteError := todo.Delete(ctx, ts.db)
	if deleteError != nil {
		return http.StatusInternalServerError, deleteError
	}
	return http.StatusOK, nil
}

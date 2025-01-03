package services

import (
	"app/generated/todos"
	models "app/models/generated"
	"app/test/factories"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type TestTodoServiceSuite struct {
	WithDBSuite
}

var (
	user            *models.User
	testTodoService TodoService
)

func (s *TestTodoServiceSuite) SetupTest() {
	s.SetDBCon()

	// NOTE: テスト用ユーザの作成
	user = factories.UserFactory.MustCreateWithOption(map[string]interface{}{"Email": "test@example.com"}).(*models.User)
	if err := user.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test user %v", err)
	}

	testTodoService = NewTodoService(DBCon)
}

func (s *TestTodoServiceSuite) TearDownTest() {
	s.CloseDB()
}

func (s *TestTodoServiceSuite) TestCreateTodo() {
	requestParams := todos.PostTodosJSONRequestBody{Title: "test title 1", Content: "test content 1"}

	statusCode, err := testTodoService.CreateTodo(ctx, requestParams, int64(user.ID))

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(http.StatusOK), statusCode)

	// NOTE: Todoリストが作成されていることを確認
	isExistTodo, _ := models.Todos(
		qm.Where("title = ?", "test title 1"),
	).Exists(ctx, DBCon)
	assert.True(s.T(), isExistTodo)
}

func (s *TestTodoServiceSuite) TestCreateTodo_ValidationError() {
	requestParams := todos.PostTodosJSONRequestBody{Title: "", Content: "test content 1"}

	statusCode, err := testTodoService.CreateTodo(ctx, requestParams, int64(user.ID))

	assert.Contains(s.T(), err.Error(), "タイトルは必須入力です。")
	assert.Equal(s.T(), int64(http.StatusBadRequest), statusCode)

	// NOTE: Todoリストが作成されていないことを確認
	isExistTodo, _ := models.Todos(
		qm.Where("title = ?", "test title 1"),
	).Exists(ctx, DBCon)
	assert.False(s.T(), isExistTodo)
}

func (s *TestTodoServiceSuite) TestFetchTodosList() {
	var todosSlice models.TodoSlice
	todosSlice = append(todosSlice, &models.Todo{
		Title:   "test title 1",
		Content: null.String{String: "test content 1", Valid: true},
		UserID:  int64(user.ID),
	})
	todosSlice = append(todosSlice, &models.Todo{
		Title:   "test title 2",
		Content: null.String{String: "test content 2", Valid: true},
		UserID:  int64(user.ID),
	})
	_, err := todosSlice.InsertAll(ctx, DBCon, boil.Infer())
	if err != nil {
		s.T().Fatalf("failed to create TestFetchTodosList Data: %v", err)
	}

	statusCode, todosList, err := testTodoService.FetchTodosList(ctx, int64(user.ID))

	assert.Equal(s.T(), int64(http.StatusOK), statusCode)
	assert.Len(s.T(), *todosList, 2)
	assert.Nil(s.T(), err)
}

func (s *TestTodoServiceSuite) TestFetchTodo_StatusOk() {
	testTodo := models.Todo{Title: "test title 1", Content: null.String{String: "test content 1", Valid: true}, UserID: int64(user.ID)}
	if err := testTodo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todos %v", err)
	}
	testTodo.Reload(ctx, DBCon)

	statusCode, todo := testTodoService.ShowTodo(ctx, testTodo.ID, int64(user.ID))

	assert.Equal(s.T(), int64(http.StatusOK), statusCode)
	assert.Equal(s.T(), testTodo.Title, todo.Title)
	assert.Equal(s.T(), testTodo.Content, todo.Content)
}

func (s *TestTodoServiceSuite) TestFetchTodo_StatusNotFound() {
	testTodo := models.Todo{Title: "test title 1", Content: null.String{String: "test content 1", Valid: true}, UserID: int64(user.ID)}
	if err := testTodo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todos %v", err)
	}
	testTodo.Reload(ctx, DBCon)

	statusCode, todo := testTodoService.ShowTodo(ctx, testTodo.ID, int64(user.ID + 1))

	assert.Equal(s.T(), int64(http.StatusNotFound), statusCode)
	assert.Equal(s.T(), "", todo.Title)
	assert.Equal(s.T(), null.String{String: "", Valid: false}, todo.Content)
}

func (s *TestTodoServiceSuite) TestUpdateTodo_StatusOk() {
	testTodo := models.Todo{Title: "test title 1", Content: null.String{String: "test content 1", Valid: true}, UserID: int64(user.ID)}
	if err := testTodo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todos %v", err)
	}

	requestParams := todos.PatchTodoJSONRequestBody{Title: "test updated title 1", Content: "test updated content 1"}
	statusCode, err := testTodoService.UpdateTodo(ctx, testTodo.ID, requestParams, int64(user.ID))

	assert.Equal(s.T(), int64(http.StatusOK), statusCode)
	assert.Nil(s.T(), err)
	// NOTE: TODOが更新されていることの確認
	if err := testTodo.Reload(ctx, DBCon); err != nil {
		s.T().Fatalf("failed to reload test todos %v", err)
	}
	assert.Equal(s.T(), "test updated title 1", testTodo.Title)
	assert.Equal(s.T(), null.String{String: "test updated content 1", Valid: true}, testTodo.Content)
}

func (s *TestTodoServiceSuite) TestUpdateTodo_ValidationError() {
	testTodo := models.Todo{Title: "test title 1", Content: null.String{String: "test content 1", Valid: true}, UserID: int64(user.ID)}
	if err := testTodo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todos %v", err)
	}

	requestParams := todos.PatchTodoJSONRequestBody{Title: "", Content: "test updated content 1"}
	statusCode, err := testTodoService.UpdateTodo(ctx, testTodo.ID, requestParams, int64(user.ID))

	assert.Contains(s.T(), err.Error(), "タイトルは必須入力です。")
	assert.Equal(s.T(), int64(http.StatusBadRequest), statusCode)
	// NOTE: Todoが更新されていないこと
	if err := testTodo.Reload(ctx, DBCon); err != nil {
		s.T().Fatalf("failed to reload test todos %v", err)
	}
	assert.Equal(s.T(), "test title 1", testTodo.Title)
	assert.Equal(s.T(), null.String{String: "test content 1", Valid: true}, testTodo.Content)
}

func (s *TestTodoServiceSuite) TestUpdateTodo_NotFound() {
	testTodo := models.Todo{Title: "test title 1", Content: null.String{String: "test content 1", Valid: true}, UserID: int64(user.ID)}
	if err := testTodo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todos %v", err)
	}

	requestParams := todos.PatchTodoJSONRequestBody{Title: "test updated title 1", Content: "test updated content 1"}
	statusCode, err := testTodoService.UpdateTodo(ctx, testTodo.ID + 1, requestParams, int64(user.ID))

	assert.Equal(s.T(), int64(http.StatusNotFound), statusCode)
	assert.NotNil(s.T(), err)
	// NOTE: TODOが更新されていないことの確認
	if err := testTodo.Reload(ctx, DBCon); err != nil {
		s.T().Fatalf("failed to reload test todos %v", err)
	}
	assert.Equal(s.T(), "test title 1", testTodo.Title)
	assert.Equal(s.T(), null.String{String: "test content 1", Valid: true}, testTodo.Content)
}

func (s *TestTodoServiceSuite) TestDeleteTodo_StatusOk() {
	testTodo := models.Todo{Title: "test title 1", Content: null.String{String: "test content 1", Valid: true}, UserID: int64(user.ID)}
	if err := testTodo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todos %v", err)
	}

	statusCode, deleteErr := testTodoService.DeleteTodo(ctx, testTodo.ID, int64(user.ID))

	assert.Equal(s.T(), int64(http.StatusOK), statusCode)
	assert.Nil(s.T(), deleteErr)
	// NOTE: TODOが削除されていることの確認
	err := testTodo.Reload(ctx, DBCon)
	assert.NotNil(s.T(), err)
}

func (s *TestTodoServiceSuite) TestDeleteTodo_NotFound() {
	testTodo := models.Todo{Title: "test title 1", Content: null.String{String: "test content 1", Valid: true}, UserID: int64(user.ID)}
	if err := testTodo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todos %v", err)
	}

	statusCode, deleteErr := testTodoService.DeleteTodo(ctx, testTodo.ID + 1, int64(user.ID))

	assert.Equal(s.T(), int64(http.StatusNotFound), statusCode)
	assert.NotNil(s.T(), deleteErr)
	// NOTE: TODOが削除されていないことの確認
	err := testTodo.Reload(ctx, DBCon)
	assert.Nil(s.T(), err)
}

func TestTodoService(t *testing.T) {
	// テストスイートを実行
	suite.Run(t, new(TestTodoServiceSuite))
}

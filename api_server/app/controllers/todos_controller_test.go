package controllers

import (
	"app/generated/auth"
	"app/generated/todos"
	"app/middlewares"
	models "app/models/generated"
	"app/services"
	"app/test/factories"
	"net/http"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/oapi-codegen/testutil"
)

var (
	testTodosController TodosController
)

type testTodosControllerSuite struct {
	WithDBSuite
}

func (s *testTodosControllerSuite) SetupTest() {
	s.SetDBCon()

	todoService := services.NewTodoService(DBCon)

	// NOTE: テスト対象のコントローラを設定
	testTodosController = NewTodosController(todoService)

	todosMiddlewares := []todos.StrictMiddlewareFunc{middlewares.AuthMiddleware}
	strictHandler := todos.NewStrictHandler(testTodosController, todosMiddlewares)
	todos.RegisterHandlers(e, strictHandler)

	authService := services.NewAuthService(DBCon)
	authController := NewAuthController(authService)

	authStrictHandler := auth.NewStrictHandler(authController, nil)
	auth.RegisterHandlers(e, authStrictHandler)
}

func (s *testTodosControllerSuite) TearDownTest() {
	s.CloseDB()
}

func (s *testTodosControllerSuite) TestPostTodos_StatusOk() {
	s.SetCsrfHeaderValues()
	s.SignIn()

	reqBody := todos.StoreTodoInput{
		Title: "test_title",
		Content: "test_content",
	}
	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Post("/todos").WithHeader("Cookie", token+"; "+csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).WithJsonBody(reqBody).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusOK, result.Code())

	var res todos.PostTodos200JSONResponse
	result.UnmarshalBodyToObject(&res)
	
	assert.Equal(s.T(), int64(http.StatusOK), res.Code)

	// NOTE: TODOリストが作成されていることを確認
	todo, err := models.Todos(
		qm.Where("title = ?", "test_title"),
	).One(ctx, DBCon)
	if err != nil {
		s.T().Fatalf("failed to create todo %v", err)
	}
	assert.Equal(s.T(), null.String{String: "test_content", Valid: true}, todo.Content)
}

func (s *testTodosControllerSuite) TestPostTodos_StatusBadRequest() {
	s.SetCsrfHeaderValues()
	s.SignIn()

	reqBody := todos.StoreTodoInput{
		Title: "",
		Content: "test_content",
	}
	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Post("/todos").WithHeader("Cookie", token+"; "+csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).WithJsonBody(reqBody).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusBadRequest, result.Code())

	var res todos.PostTodos400JSONResponse
	result.UnmarshalBodyToObject(&res)
	titleValidationErrors := *res.Errors.Title
	assert.Equal(s.T(), []string{"タイトルは必須入力です。"}, titleValidationErrors)
	
	assert.Equal(s.T(), int64(http.StatusBadRequest), res.Code)

	// NOTE: TODOリストが作成されていないことを確認
	isExistTodo, _ := models.Todos(
		qm.Where("title = ?", ""),
	).Exists(ctx, DBCon)
	assert.False(s.T(), isExistTodo)
}

func (s *testTodosControllerSuite) TestPostTodos_StatusUnauthorized() {
	s.SetCsrfHeaderValues()
	reqBody := todos.StoreTodoInput{
		Title: "test_title",
		Content: "test_content",
	}
	result := testutil.NewRequest().Post("/todos").WithHeader("Cookie", csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).WithJsonBody(reqBody).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusUnauthorized, result.Code())

	// NOTE: TODOリストが作成されていないことを確認
	isExistTodo, _ := models.Todos(
		qm.Where("title = ?", "test_title"),
	).Exists(ctx, DBCon)
	assert.False(s.T(), isExistTodo)
}

func (s *testTodosControllerSuite) TestGetTodos_StatusOk() {
	s.SetCsrfHeaderValues()
	s.SignIn()

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
	
	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Get("/todos").WithHeader("Cookie", token+"; "+csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusOK, result.Code())

	var res todos.GetTodos200JSONResponse
	result.UnmarshalBodyToObject(&res)

	assert.Equal(s.T(), 2, len(res.Todos))
	assert.Equal(s.T(), "test title 1", res.Todos[0].Title)
	assert.Equal(s.T(), "test content 1", res.Todos[0].Content)
}

func (s *testTodosControllerSuite) TestGetTodos_StatusUnauthorized() {
	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Get("/todos").WithHeader("Cookie", csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusUnauthorized, result.Code())
}

func (s *testTodosControllerSuite) TestGetTodo_StatusOk() {
	s.SetCsrfHeaderValues()
	s.SignIn()

	todoParam := map[string]interface{}{"UserID": int64(user.ID), "Title": "test title 1", "Content": null.String{String: "test content 1", Valid: true}}
	todo := factories.TodoFactory.MustCreateWithOption(todoParam).(*models.Todo)
	if err := todo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todo %v", err)
	}

	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Get("/todos/"+strconv.Itoa(int(todo.ID))).WithHeader("Cookie", token+"; "+csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusOK, result.Code())

	var res todos.GetTodo200JSONResponse
	result.UnmarshalBodyToObject(&res)

	assert.Equal(s.T(), "test title 1", res.Todo.Title)
	assert.Equal(s.T(), "test content 1", res.Todo.Content)
}

func (s *testTodosControllerSuite) TestGetTodo_StatusUnauthorized() {
	s.SetCsrfHeaderValues()
	
	result := testutil.NewRequest().Get("/todos/1").WithHeader("Cookie", csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusUnauthorized, result.Code())
}

func (s *testTodosControllerSuite) TestGetTodo_StatusNotFound() {
	s.SetCsrfHeaderValues()
	s.SignIn()

	todoParam := map[string]interface{}{"UserID": int64(user.ID), "Title": "test title 1", "Content": null.String{String: "test content 1", Valid: true}}
	todo := factories.TodoFactory.MustCreateWithOption(todoParam).(*models.Todo)
	if err := todo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todo %v", err)
	}
	
	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Get("/todos/"+strconv.Itoa(int(todo.ID + 1))).WithHeader("Cookie", token+"; "+csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusNotFound, result.Code())
}

func (s *testTodosControllerSuite) TestPatchTodo_StatusOk() {
	s.SetCsrfHeaderValues()
	s.SignIn()

	todoParam := map[string]interface{}{"UserID": int64(user.ID), "Title": "test title 1", "Content": null.String{String: "test content 1", Valid: true}}
	todo := factories.TodoFactory.MustCreateWithOption(todoParam).(*models.Todo)
	if err := todo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todo %v", err)
	}

	reqBody := todos.StoreTodoInput{
		Title: "test updated title 1",
		Content: "test updated content 1",
	}
	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Patch("/todos/"+strconv.Itoa(int(todo.ID))).WithHeader("Cookie", token+"; "+csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).WithJsonBody(reqBody).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusOK, result.Code())

	var res todos.PatchTodo200JSONResponse
	result.UnmarshalBodyToObject(&res)
	
	assert.Equal(s.T(), int64(http.StatusOK), res.Code)

	// NOTE: TODOリストが更新されていることを確認
	if err := todo.Reload(ctx, DBCon); err != nil {
		s.T().Fatalf("failed to reload test todos %v", err)
	}
	assert.Equal(s.T(), "test updated title 1", todo.Title)
	assert.Equal(s.T(), null.String{String: "test updated content 1", Valid: true}, todo.Content)
}

func (s *testTodosControllerSuite) TestPatchTodo_StatusBadRequest() {
	s.SetCsrfHeaderValues()
	s.SignIn()

	todoParam := map[string]interface{}{"UserID": int64(user.ID), "Title": "test title 1", "Content": null.String{String: "test content 1", Valid: true}}
	todo := factories.TodoFactory.MustCreateWithOption(todoParam).(*models.Todo)
	if err := todo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todo %v", err)
	}
	
	reqBody := todos.StoreTodoInput{
		Title: "",
		Content: "test updated content 1",
	}
	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Patch("/todos/"+strconv.Itoa(int(todo.ID))).WithHeader("Cookie", token+"; "+csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).WithJsonBody(reqBody).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusBadRequest, result.Code())

	var res todos.PatchTodo400JSONResponse
	result.UnmarshalBodyToObject(&res)
	titleValidationErrors := *res.Errors.Title
	assert.Equal(s.T(), []string{"タイトルは必須入力です。"}, titleValidationErrors)
	
	assert.Equal(s.T(), int64(http.StatusBadRequest), res.Code)

	// NOTE: TODOリストが更新されていないことを確認
	if err := todo.Reload(ctx, DBCon); err != nil {
		s.T().Fatalf("failed to reload test todos %v", err)
	}
	assert.Equal(s.T(), "test title 1", todo.Title)
	assert.Equal(s.T(), null.String{String: "test content 1", Valid: true}, todo.Content)
}

func (s *testTodosControllerSuite) TestPatchTodo_StatusUnauthorized() {
	s.SetCsrfHeaderValues()

	reqBody := todos.StoreTodoInput{
		Title: "test_title",
		Content: "test_content",
	}
	result := testutil.NewRequest().Patch("/todos/1").WithHeader("Cookie", csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).WithJsonBody(reqBody).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusUnauthorized, result.Code())
}

func (s *testTodosControllerSuite) TestPatchTodo_StatusNotFound() {
	s.SetCsrfHeaderValues()
	s.SignIn()

	todoParam := map[string]interface{}{"UserID": int64(user.ID), "Title": "test title 1", "Content": null.String{String: "test content 1", Valid: true}}
	todo := factories.TodoFactory.MustCreateWithOption(todoParam).(*models.Todo)
	if err := todo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todo %v", err)
	}
	
	reqBody := todos.StoreTodoInput{
		Title: "test updated title 1",
		Content: "test updated content 1",
	}
	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Patch("/todos/"+strconv.Itoa(int(todo.ID + 1))).WithHeader("Cookie", token+"; "+csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).WithJsonBody(reqBody).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusNotFound, result.Code())

	// NOTE: TODOリストが更新されていないことを確認
	if err := todo.Reload(ctx, DBCon); err != nil {
		s.T().Fatalf("failed to reload test todos %v", err)
	}
	assert.Equal(s.T(), "test title 1", todo.Title)
	assert.Equal(s.T(), null.String{String: "test content 1", Valid: true}, todo.Content)
}

func (s *testTodosControllerSuite) TestDeleteTodo_StatusOk() {
	s.SetCsrfHeaderValues()
	s.SignIn()

	todoParam := map[string]interface{}{"UserID": int64(user.ID), "Title": "test title 1", "Content": null.String{String: "test content 1", Valid: true}}
	todo := factories.TodoFactory.MustCreateWithOption(todoParam).(*models.Todo)
	if err := todo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todo %v", err)
	}

	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Delete("/todos/"+strconv.Itoa(int(todo.ID))).WithHeader("Cookie", token+"; "+csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusOK, result.Code())

	var res todos.DeleteTodo200JSONResponse
	result.UnmarshalBodyToObject(&res)
	
	assert.Equal(s.T(), int64(http.StatusOK), res.Code)
	assert.Equal(s.T(), true, res.Result)

	// NOTE: TODOリストが削除されていることを確認
	err := todo.Reload(ctx, DBCon)
	assert.NotNil(s.T(), err)
}

func (s *testTodosControllerSuite) TestDeleteTodo_StatusUnauthorized() {
	s.SetCsrfHeaderValues()

	result := testutil.NewRequest().Delete("/todos/1").WithHeader("Cookie", csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusUnauthorized, result.Code())
}

func (s *testTodosControllerSuite) TestDeleteTodo_StatusNotFound() {
	s.SetCsrfHeaderValues()
	s.SignIn()

	todoParam := map[string]interface{}{"UserID": int64(user.ID), "Title": "test title 1", "Content": null.String{String: "test content 1", Valid: true}}
	todo := factories.TodoFactory.MustCreateWithOption(todoParam).(*models.Todo)
	if err := todo.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test todo %v", err)
	}
	
	s.SetCsrfHeaderValues()
	result := testutil.NewRequest().Delete("/todos/"+strconv.Itoa(int(todo.ID + 1))).WithHeader("Cookie", token+"; "+csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).GoWithHTTPHandler(s.T(), e)
	assert.Equal(s.T(), http.StatusNotFound, result.Code())

	// NOTE: TODOリストが削除されていないことを確認
	err := todo.Reload(ctx, DBCon)
	assert.Nil(s.T(), err)
}

func TestTodosController(t *testing.T) {
	// テストスイートを実施
	suite.Run(t, new(testTodosControllerSuite))
}

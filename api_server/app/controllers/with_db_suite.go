package controllers

import (
	"app/db"
	"app/generated/auth"
	models "app/models/generated"
	"app/test/factories"
	"app/utils/routers"
	"context"
	"database/sql"

	"github.com/DATA-DOG/go-txdb"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/testutil"
	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type WithDBSuite struct {
	suite.Suite
}

var (
	DBCon *sql.DB
	ctx   context.Context
	user *models.User
	token string
	e *echo.Echo
	csrfToken string
	csrfTokenCookie string
)

// func (s *WithDBSuite) SetupSuite()                           {} // テストスイート実施前の処理
// func (s *WithDBSuite) TearDownSuite()                        {} // テストスイート終了後の処理
// func (s *WithDBSuite) SetupTest()                            {} // テストケース実施前の処理
// func (s *WithDBSuite) TearDownTest()                         {} // テストケース終了後の処理
// func (s *WithDBSuite) BeforeTest(suiteName, testName string) {} // テストケース実施前の処理
// func (s *WithDBSuite) AfterTest(suiteName, testName string)  {} // テストケース終了後の処理

func init() {
	txdb.Register("txdb-controller", "mysql", db.GetDsn())
	ctx = context.Background()

	e = routers.ApplyMiddlewares(echo.New())
}

func (s *WithDBSuite) SetDBCon() {
	db, err := sql.Open("txdb-controller", "connect")
	if err != nil {
		s.T().Fatalf("failed to initialize DB: %v", err)
	}
	DBCon = db
}

func (s *WithDBSuite) CloseDB() {
	DBCon.Close()
}

func (s *WithDBSuite) SignIn() {
	// NOTE: テスト用ユーザの作成
	user = factories.UserFactory.MustCreateWithOption(map[string]interface{}{"Email": "test@example.com"}).(*models.User)
	if err := user.Insert(ctx, DBCon, boil.Infer()); err != nil {
		s.T().Fatalf("failed to create test user %v", err)
	}
	
	reqBody := auth.SignInInput{
		Email: "test@example.com",
		Password: "password",
	}

	closed := make(chan struct{})
	go func() {
		result := testutil.NewRequest().Post("/auth/signIn").WithHeader("Cookie", csrfTokenCookie).WithHeader(echo.HeaderXCSRFToken, csrfToken).WithJsonBody(reqBody).GoWithHTTPHandler(s.T(), e)

		token = result.Recorder.Result().Header.Values("Set-Cookie")[0]	
		close(closed)
	}()

	<-closed
}

func (s *WithDBSuite) SetCsrfHeaderValues() {
	closed := make(chan struct{})
	go func ()  {
		result := testutil.NewRequest().Get("/auth/csrf").GoWithHTTPHandler(s.T(), e)

		var res auth.GetAuthCsrf200JSONResponse
		err := result.UnmarshalJsonToObject(&res)
		if err != nil {
			s.T().Error(err.Error())
		}

		csrfToken = res.CsrfToken
		csrfTokenCookie = result.Recorder.Result().Header.Values("Set-Cookie")[0]

		close(closed)
	}()

	<-closed
}

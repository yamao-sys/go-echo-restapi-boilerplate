package controllers

import (
	"app/db"
	"context"
	"database/sql"

	"github.com/DATA-DOG/go-txdb"
	"github.com/stretchr/testify/suite"
)

type WithDBSuite struct {
	suite.Suite
}

var (
	DBCon *sql.DB
	ctx   context.Context
	// token string
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

// func (s *WithDBSuite) SignIn() {
// 	authService := services.NewAuthService(DBCon)
// 	authController := NewAuthController(authService)

// 	// recorderの初期化
// 	authRecorder := httptest.NewRecorder()

// 	// NOTE: リクエストの生成
// 	signUpRequestBody := bytes.NewBufferString("{\"name\":\"test name 1\",\"email\":\"test@example.com\",\"password\":\"password\"}")
// 	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", signUpRequestBody)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

// 	// echoによるWebサーバの設定
// 	echoServer := echo.New()
// 	c := echoServer.NewContext(req, authRecorder)
// 	c.SetPath("auth/sign_up")

// 	// NOTE: ログインし、tokenに認証情報を格納
// 	authController.SignIn(c)
// 	token = authRecorder.Result().Cookies()[0].Value
// }

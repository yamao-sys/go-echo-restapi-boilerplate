package services

import (
	"app/db"
	"context"
	"database/sql"

	"github.com/DATA-DOG/go-txdb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/suite"
)

type WithDBSuite struct {
	suite.Suite
}

var DBCon *sql.DB
var ctx context.Context

// func (s *WithDBSuite) SetupSuite()                           {} // テストスイート実施前の処理
// func (s *WithDBSuite) TearDownSuite()                        {} // テストスイート終了後の処理
// func (s *WithDBSuite) SetupTest()                            {} // テストケース実施前の処理
// func (s *WithDBSuite) TearDownTest()                         {} // テストケース終了後の処理
// func (s *WithDBSuite) BeforeTest(suiteName, testName string) {} // テストケース実施前の処理
// func (s *WithDBSuite) AfterTest(suiteName, testName string)  {} // テストケース終了後の処理

func init() {
	txdb.Register("txdb-service", "mysql", db.GetDsn())
	ctx = context.Background()
}

func (s *WithDBSuite) SetDBCon() {
	db, err := sql.Open("txdb-service", "connect")
	if err != nil {
		s.T().Fatalf("failed to initialize DB: %v", err)
	}
	DBCon = db
}

func (s *WithDBSuite) CloseDB() {
	DBCon.Close()
}

package db

import (
	"database/sql"
	"log"
	"os"
)

func Init() *sql.DB {
	// DBインスタンス生成
	db, err := sql.Open("mysql", GetDsn())
	if err != nil {
		log.Fatalln(err)
	}
	return db
}

func Close(db *sql.DB) {
	if err := db.Close(); err != nil {
		panic(err)
	}
}

func GetDsn() string {
	return os.Getenv("MYSQL_USER") +
		":" + os.Getenv("MYSQL_PASS") +
		"@tcp(" + os.Getenv("MYSQL_HOST") + ":" + os.Getenv("MYSQL_PORT") + ")/" +
		os.Getenv("MYSQL_DBNAME") +
		"?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local"
}

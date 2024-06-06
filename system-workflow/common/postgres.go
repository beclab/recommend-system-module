package common

import (
	"database/sql"
	"log"
	"os"
)

func DbOpen() *sql.DB {
	connStr := os.Getenv("TERMINUS_RECOMMEND_POSTGRES_CONN")
	if connStr == "" {
		connStr = "host=localhost port=5432 user=angelhand password=2222 dbname=ahdb sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

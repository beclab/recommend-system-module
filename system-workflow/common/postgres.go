package common

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	// Postgresql driver import
	_ "github.com/lib/pq"
)

func GetPGHost() string {
	host := os.Getenv("PG_HOST")
	if host == "" {
		return "127.0.0.1"
	}
	return host
}

func GetPGUser() string {
	user := os.Getenv("PG_USERNAME")
	if user == "" {
		return "postgres"
	}
	return user

}

func GetPGPass() string {
	pass := os.Getenv("PG_PASSWORD")
	if pass == "" {
		return "postgres"
	}
	return pass
}

func GetPGDbName() string {
	dbName := os.Getenv("PG_DATABASE")
	if dbName == "" {
		return "rss"
	}
	return dbName
}

func GetPGPort() int {
	return ParseInt(os.Getenv("PG_PORT"), 5432)
}

func GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", GetPGHost(), GetPGPort(), GetPGUser(), GetPGPass(), GetPGDbName())
}

func NewPostgresClient() *sql.DB {
	connStr := GetDatabaseURL()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

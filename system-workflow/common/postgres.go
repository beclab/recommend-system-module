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
		return "124.222.40.95"
	}
	return host
}

func GetPGUser() string {
	user := os.Getenv("PG_USER")
	if user == "" {
		return "postgres"
	}
	return user

}

func GetPGPass() string {
	pass := os.Getenv("PG_PASS")
	if pass == "" {
		return "liujx123"
	}
	return pass
}

func GetPGDbName() string {
	dbName := os.Getenv("PG_DB_NAME")
	if dbName == "" {
		return "rss_v4"
	}
	return dbName
}
func GetDatabaseURL() string {
	return fmt.Sprintf("host=%s  user=%s password=%s dbname=%s sslmode=disable", GetPGHost(), GetPGUser(), GetPGPass(), GetPGDbName())
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

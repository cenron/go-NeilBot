package storage

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type Storage struct {
	DB *sqlx.DB
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{
		DB: db,
	}
}

func NewDB(filepath string) *sqlx.DB {
	if _, err := os.Stat(filepath); err != nil {
		if _, err := os.Create(filepath); err != nil {
			log.Fatal("could not open database")
		}
	}

	db, err := sqlx.Connect("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

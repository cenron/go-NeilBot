package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/cenron/neil-bot-go/pkg/util"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type Storage struct {
	DB *sql.DB
}

func NewStorage(filepath string) *Storage {

	util.LoadEnv()

	log.Print(filepath)
	if _, err := os.Stat(filepath); err != nil {
		os.Create(filepath)
	}

	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}

	return &Storage{
		DB: db,
	}
}

func CreateTable(db *sql.DB) {
	createStudentTableSQL := `CREATE TABLE student (
		"idStudent" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"code" TEXT,
		"name" TEXT,
		"program" TEXT		
	  );` // SQL Statement for Create Table

	log.Println("Create student table...")
	statement, err := db.Prepare(createStudentTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("student table created")
}

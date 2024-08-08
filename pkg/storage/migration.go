package storage

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"slices"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigration(db *sqlx.DB) {

	// First try to read our meta data of existing migrations.
	metadata := readMetaData(db)

	// Find a list migration scripts to run
	migrations := GetListMigration()

	// Clean our up migrations that have been applied already.
	files := removeDuplicate(metadata, migrations)
	if len(files) == 0 {
		log.Print("no migrations to run. skipping for now")
		return
	}

	// Sort our files to make sure they are in the right order
	sort.Strings(files)

	for _, f := range files {
		log.Printf("Running migration file: %s", f)

		query := ReadMigration(f)
		if query == "" {
			continue
		}

		_, err := db.Exec(query)
		if err != nil {
			log.Printf("could not execute migration: %s -> %v", f, err)
			continue
		}

		// Generate hash for our file.
		hash, err := hashFile(query)
		if err != nil {
			log.Printf("could not find hash for file: %v", err)
		}

		// Add to our migration table
		db.Exec("INSERT INTO migration (name, hash, created_at, updated_at) VALUES($1, $2, $3, $4)", f, hash, time.Now(), time.Now())
	}
}

func readMetaData(db *sqlx.DB) []string {
	var migrations []string

	// We need to get our migration metadata, otherwise create it and return an empty list.
	rows, err := db.Query("SELECT name FROM migration")
	if err, ok := err.(sqlite3.Error); ok {
		if err.Code == sqlite3.ErrError {
			log.Print("could not migration table. initializing table.")
			db.Exec("CREATE TABLE IF NOT EXISTS migration (migration_id INTEGER PRIMARY KEY AUTOINCREMENT, name VARCHAR(255), hash VARCHAR(64) UNIQUE, created_at TIMESTAMP, updated_at TIMESTAMP)")
			return migrations
		}

		log.Fatalf("error running migration: %v", err.Code)
	}

	// Populate our migrations with the data from the migration table.
	defer rows.Close()
	for rows.Next() {
		var migration string
		err := rows.Scan(&migration)
		if err != nil {
			log.Print("error while trying to read migrations")
			return migrations
		}

		migrations = append(migrations, migration)
	}

	return migrations
}

func removeDuplicate(metadata []string, migrations []string) []string {
	var cleanMigrations []string
	for _, file := range migrations {
		if slices.Contains(metadata, file) {
			continue
		}

		// Append to our clean list.
		cleanMigrations = append(cleanMigrations, file)
	}

	return cleanMigrations
}

func hashFile(data string) (string, error) {
	// Get the file hash
	h := sha256.New()
	if _, err := h.Write([]byte(data)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func GetListMigration() []string {

	var files []string
	err := fs.WalkDir(migrationsFS, "migrations", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil
	}

	return files
}

func ReadMigration(fileName string) string {
	data, err := fs.ReadFile(migrationsFS, fileName)
	if err != nil {
		return ""
	}

	return string(data)
}

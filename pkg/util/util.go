package util

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// use to return error for use with slog
func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

// exists returns whether the given file or directory exists
func DirExists(path string) bool {
	f, err := os.Stat(path)
	if err == nil {
		return f.IsDir()
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// loads the env variables.
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

package os

import (
	"errors"
	"os"
)

func Exists(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

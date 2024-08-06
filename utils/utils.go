package utils

import (
	"log"
	"os"
)

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func DeleteFileIfExists(filename string) {
	if FileExists(filename) {
		err := os.Remove(filename)
		if err != nil {
			log.Fatalf("Error deleting file %v: %v", filename, err)
		}
	}
}

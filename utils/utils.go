package utils

import (
	"io/ioutil"
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

func CopyFile(sourceFile string, destFile string) {
	fin, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Fatalf("Error opening file for copying: %v", err)
	}

	err = ioutil.WriteFile(destFile, fin, 0644)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}
}

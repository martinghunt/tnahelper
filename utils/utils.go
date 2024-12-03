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

func RenameFile(oldName string, newName string) {
	err := os.Rename(oldName, newName)
	if err != nil {
		log.Fatalf("Error renaming file %s -> %s: %v", oldName, newName, err)
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

func ReverseComplement(seq []byte) []byte {
	revcomp := make([]byte, len(seq))

	comp := map[byte]byte{
		'A': 'T',
		'T': 'A',
		'C': 'G',
		'G': 'C',
		'N': 'N',
	}

	for i := 0; i < len(seq); i++ {
		revcomp[len(seq)-1-i] = comp[seq[i]]
	}

	return revcomp
}

func Reverse(s string) string {
	rev := []byte(s)
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return string(rev)
}

package utils

import (
	"github.com/shenwei356/xopen"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	require.False(t, FileExists("this_file_does_not_exist"), "File exists but should not exist")
	filename := filepath.Join("utils_testdata", "fileExists")
	require.True(t, FileExists(filename), "File not found, but should have found it: 'fileExists'")
}

func TestDeleteFileIfExists(t *testing.T) {
	outfile := "tmp.test.DeleteFileIfExists"
	DeleteFileIfExists(outfile)
	require.False(t, FileExists(outfile), "File should not exist: %v", outfile)
	fout, errOut := xopen.Wopen(outfile)
	require.Equal(t, errOut, nil, "Error opening test file for writing: %v", outfile)
	fout.Close()
	require.True(t, FileExists(outfile), "File should exist: %v", outfile)
	DeleteFileIfExists(outfile)
	require.False(t, FileExists(outfile), "File should not exist: %v", outfile)
	DeleteFileIfExists(outfile) // run again, should not throw error
}

func TestCopyFile(t *testing.T) {
	infile := filepath.Join("utils_testdata", "copyFile")
	outfile := "tmp.test.CopyFile.out"
	DeleteFileIfExists(outfile)
	CopyFile(infile, outfile)
	require.True(t, FileExists(outfile), "File should exist: %v", outfile)
	DeleteFileIfExists(outfile)
}

func TestReverseComplement(t *testing.T) {
	seq := []byte("ACCGTN")
	expect := []byte("NACGGT")
	rev := ReverseComplement(seq)
	require.Equal(t, string(rev), string(expect), "Error reverse complement. Got: %s", rev)
}

func TestReverse(t *testing.T) {
	fwd := "ABCDE"
	rev := "EDCBA"
	got := Reverse(fwd)
	require.Equal(t, rev, got, "Error Reverse. Got: %s", got)
}

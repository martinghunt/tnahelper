package example_data

import (
	"github.com/martinghunt/tnahelper/utils"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestMakeTestData(t *testing.T) {
	outdir := "tmp.test_data"
	genome1_gff := filepath.Join(outdir, "g1.gff")
	genome2_gff := filepath.Join(outdir, "g2.gff")
	utils.DeleteFileIfExists(genome1_gff)
	utils.DeleteFileIfExists(genome2_gff)
	utils.DeleteFileIfExists(outdir)
	MakeTestFiles(outdir)
	require.True(t, utils.FileExists(genome1_gff), "File not found, but should have found it: %v", genome1_gff)
	utils.DeleteFileIfExists(genome1_gff)
	utils.DeleteFileIfExists(genome2_gff)
	utils.DeleteFileIfExists(outdir)
}

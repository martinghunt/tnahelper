package download

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
	"github.com/martinghunt/tnahelper/utils"
)

func TestFastaAndGffFromZipOK(t *testing.T) {
	zipfile := filepath.Join("download_testdata", "fastaAndGffFromZip_ok.zip")
	outprefix := "tmp.TestFastaAndGffFromZip_ok"
	fa := outprefix + ".fa"
	gff := outprefix + ".gff"
	utils.DeleteFileIfExists(fa)
	utils.DeleteFileIfExists(gff)
	FastaAndGffFromZip(zipfile, outprefix)
	require.True(t, utils.FileExists(fa), "FASTA file not found %s", fa)
	require.True(t, utils.FileExists(gff), "GFF file not found %s", gff)
	utils.DeleteFileIfExists(fa)
	utils.DeleteFileIfExists(gff)
}


func TestFastaAndGffFromZipNoGFF(t *testing.T) {
	zipfile := filepath.Join("download_testdata", "fastaAndGffFromZip_no_gff.zip")
	outprefix := "tmp.TestFastaAndGffFromZip_no_gff"
	fa := outprefix + ".fa"
	gff := outprefix + ".gff"
	utils.DeleteFileIfExists(fa)
	utils.DeleteFileIfExists(gff)
	FastaAndGffFromZip(zipfile, outprefix)
	require.True(t, utils.FileExists(fa), "FASTA file not found %s", fa)
	require.False(t, utils.FileExists(gff), "GFF file found %s", gff)
	utils.DeleteFileIfExists(fa)
}


func TestFastaAndGffFromZipNoFasta(t *testing.T) {
	zipfile := filepath.Join("download_testdata", "fastaAndGffFromZip_no_fasta.zip")
	outprefix := "tmp.TestFastaAndGffFromZip_no_fasta"
	fa := outprefix + ".fa"
	gff := outprefix + ".gff"
	utils.DeleteFileIfExists(fa)
	utils.DeleteFileIfExists(gff)
	FastaAndGffFromZip(zipfile, outprefix)
	require.False(t, utils.FileExists(fa), "FASTA file found %s", fa)
	require.False(t, utils.FileExists(gff), "GFF file found %s", gff)
}


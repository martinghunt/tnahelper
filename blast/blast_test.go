package blast

import (
	"github.com/martinghunt/tnahelper/utils"
	"github.com/stretchr/testify/require"
	"github.com/udhos/equalfile"
	"path/filepath"
	"testing"
)

func TestParseBlastnFile(t *testing.T) {
	infile := filepath.Join("blast_testdata", "parse_blastn.in")
	outfile := "tmp.test.ParseBlastn"
	utils.DeleteFileIfExists(outfile)
	ParseBlastFile(infile, outfile, "blastn")
	expectFile := filepath.Join("blast_testdata", "parse_blastn.expect")
	cmp := equalfile.New(nil, equalfile.Options{})
	filesEqual, err := cmp.CompareFile(expectFile, outfile)
	require.NoError(t, err, "Error comparing blastn files %s, %s", expectFile, outfile)
	require.True(t, filesEqual, "blastn file %s expected contents incorrect", outfile)
	utils.DeleteFileIfExists(outfile)
}

func TestParseTblastxFile(t *testing.T) {
	infile := filepath.Join("blast_testdata", "parse_tblastx.in")
	outfile := "tmp.test.ParseTblastx"
	utils.DeleteFileIfExists(outfile)
	ParseBlastFile(infile, outfile, "tblastx")
	expectFile := filepath.Join("blast_testdata", "parse_tblastx.expect")
	cmp := equalfile.New(nil, equalfile.Options{})
	filesEqual, err := cmp.CompareFile(expectFile, outfile)
	require.NoError(t, err, "Error comparing tblastx files %s, %s", expectFile, outfile)
	require.True(t, filesEqual, "tlastx file %s expected contents incorrect", outfile)
	utils.DeleteFileIfExists(outfile)
}

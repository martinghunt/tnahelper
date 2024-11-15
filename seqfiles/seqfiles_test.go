package seqfiles

import (
	"github.com/martinghunt/tnahelper/utils"
	"github.com/stretchr/testify/require"
	"github.com/udhos/equalfile"
	"path/filepath"
	"testing"
)

func TestParseFASTA(t *testing.T) {
	infile := filepath.Join("seqfiles_testdata", "parseFasta.in.fa.gz")
	fileType := getFileType(infile)
	require.Equal(t, FASTA, fileType, "Did not get filetype of FASTA")
	outprefix := "tmp.test.ParseFasta"
	outfile := outprefix + ".fa"
	utils.DeleteFileIfExists(outfile)
	ParseSeqFile(infile, outprefix)
	expectFile := filepath.Join("seqfiles_testdata", "parseFasta.expect.fa")
	cmp := equalfile.New(nil, equalfile.Options{})
	filesEqual, err := cmp.CompareFile(expectFile, outfile)
	require.NoError(t, err, "Error comparing FASTA files %s, %s", expectFile, outfile)
	require.True(t, filesEqual, "FASTA file %s expected contents incorrect", outfile)
	utils.DeleteFileIfExists(outfile)

	infile = filepath.Join("seqfiles_testdata", "parseFasta.no_final_newline.in.fa")
	expectFile = filepath.Join("seqfiles_testdata", "parseFasta.no_final_newline.expect.fa")
	ParseSeqFile(infile, outprefix)
	filesEqual, err = cmp.CompareFile(expectFile, outfile)
	require.NoError(t, err, "Error comparing FASTA files %s, %s", expectFile, outfile)
	require.True(t, filesEqual, "FASTA file %s expected contents incorrect", outfile)
	utils.DeleteFileIfExists(outfile)

}

func TestParseFASTQ(t *testing.T) {
	infile := filepath.Join("seqfiles_testdata", "parseFastq.in.fq")
	fileType := getFileType(infile)
	require.Equal(t, FASTQ, fileType, "Did not get filetype of FASTQ")
	outprefix := "tmp.test.ParseFastq"
	outfile := outprefix + ".fa"
	utils.DeleteFileIfExists(outfile)
	ParseSeqFile(infile, outprefix)
	expectFile := filepath.Join("seqfiles_testdata", "parseFastq.expect.fa")
	cmp := equalfile.New(nil, equalfile.Options{})
	filesEqual, err := cmp.CompareFile(expectFile, outfile)
	require.NoError(t, err, "Error comparing FASTA files %s, %s", expectFile, outfile)
	require.True(t, filesEqual, "FASTA file %s expected contents incorrect", outfile)
	utils.DeleteFileIfExists(outfile)
}

func TestParseGFF3(t *testing.T) {
	infile := filepath.Join("seqfiles_testdata", "parseGFF3.in.gff")
	fileType := getFileType(infile)
	require.Equal(t, GFF3, fileType, "Did not get filetype of GFF3")
	outprefix := "tmp.test.ParseGFF3"
	outfileFa := outprefix + ".fa"
	utils.DeleteFileIfExists(outfileFa)
	outfileAnnot := outprefix + ".gff"
	utils.DeleteFileIfExists(outfileAnnot)
	ParseSeqFile(infile, outprefix)

	expectFileFa := filepath.Join("seqfiles_testdata", "parseGFF3.expect.fa")
	cmp := equalfile.New(nil, equalfile.Options{})
	filesEqual, err := cmp.CompareFile(expectFileFa, outfileFa)
	require.NoError(t, err, "Error comparing FASTA files %s, %s", expectFileFa, outfileFa)
	require.True(t, filesEqual, "FASTA file %s expected contents incorrect", outfileFa)
	utils.DeleteFileIfExists(outfileFa)

	expectFileAnnot := filepath.Join("seqfiles_testdata", "parseGFF3.expect.gff")
	filesEqual, err = cmp.CompareFile(expectFileAnnot, outfileAnnot)
	require.NoError(t, err, "Error comparing annotation files %s, %s", expectFileFa, outfileFa)
	require.True(t, filesEqual, "Annotation file %s expected contents incorrect", outfileFa)
	utils.DeleteFileIfExists(outfileAnnot)
}

func TestParseGenbank(t *testing.T) {
	infile := filepath.Join("seqfiles_testdata", "parseGenbank.in.gbk")
	fileType := getFileType(infile)
	require.Equal(t, GENBANK, fileType, "Did not get filetype of GENBANK")
	outprefix := "tmp.test.ParseGenbank"
	outfileFa := outprefix + ".fa"
	utils.DeleteFileIfExists(outfileFa)
	outfileAnnot := outprefix + ".gff"
	utils.DeleteFileIfExists(outfileAnnot)
	ParseSeqFile(infile, outprefix)

	expectFileFa := filepath.Join("seqfiles_testdata", "parseGenbank.expect.fa")
	cmp := equalfile.New(nil, equalfile.Options{})
	filesEqual, err := cmp.CompareFile(expectFileFa, outfileFa)
	require.NoError(t, err, "Error comparing FASTA files %s, %s", expectFileFa, outfileFa)
	require.True(t, filesEqual, "FASTA file %s expected contents incorrect", outfileFa)
	utils.DeleteFileIfExists(outfileFa)

	expectFileAnnot := filepath.Join("seqfiles_testdata", "parseGenbank.expect.gff")
	filesEqual, err = cmp.CompareFile(expectFileAnnot, outfileAnnot)
	require.NoError(t, err, "Error comparing annotation files %s, %s", expectFileFa, outfileFa)
	require.True(t, filesEqual, "Annotation file %s expected contents incorrect", outfileFa)
	utils.DeleteFileIfExists(outfileAnnot)
}

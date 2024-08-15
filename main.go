package main

import (
	"github.com/martinghunt/tnahelper/blast"
	"github.com/martinghunt/tnahelper/download"
	"github.com/martinghunt/tnahelper/seqfiles"
	"github.com/spf13/cobra"
)

var Version = "development"

func main() {
	rootCmd := &cobra.Command{Use: "tnahelper"}
	rootCmd.Version = Version
	var infile string
	var outprefix string
	var outdir string
	var bindir string

	var cmdImportSeqfile = &cobra.Command{
		Use:   "import_seqfile",
		Short: "Import sequence file",
		Run: func(cmd *cobra.Command, args []string) {
			seqfiles.ParseSeqFile(infile, outprefix)
		},
	}

	cmdImportSeqfile.Flags().StringVarP(&infile, "infile", "i", "", "REQUIRED. Input sequence file")
	cmdImportSeqfile.Flags().StringVarP(&outprefix, "outprefix", "o", "", "REQUIRED. Prefix of output files")
	cmdImportSeqfile.MarkFlagRequired("infile")
	cmdImportSeqfile.MarkFlagRequired("outprefix")
	rootCmd.AddCommand(cmdImportSeqfile)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	var cmdDownloadBinaries = &cobra.Command{
		Use:   "download_binaries",
		Short: "Download binary files",
		Run: func(cmd *cobra.Command, args []string) {
			download.DownloadBinaries(outdir)
		},
	}
	cmdDownloadBinaries.Flags().StringVarP(&outdir, "outdir", "o", "", "REQUIRED. Output directory")
	cmdDownloadBinaries.MarkFlagRequired("outdir")
	rootCmd.AddCommand(cmdDownloadBinaries)

	var cmdBlast = &cobra.Command{
		Use:   "blast",
		Short: "Run makeblastdb and blastn",
		Run: func(cmd *cobra.Command, args []string) {
			blast.RunBlast(outdir, bindir)
		},
	}
	cmdBlast.Flags().StringVarP(&outdir, "outdir", "o", "", "REQUIRED. Output directory. Must already exist and have fasta giles g1.fa,g2.fa")
	cmdBlast.Flags().StringVarP(&bindir, "bindir", "b", "", "REQUIRED. Bin directory, must contain makeblastdb,blastn")
	cmdBlast.MarkFlagRequired("outdir")
	cmdBlast.MarkFlagRequired("bindir")
	rootCmd.AddCommand(cmdBlast)

	rootCmd.Execute()
}

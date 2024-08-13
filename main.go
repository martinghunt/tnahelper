package main

import (
	"github.com/martinghunt/tnahelper/download"
	"github.com/martinghunt/tnahelper/seqfiles"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "tnahelper"}
	var infile string
	var outprefix string
	var outdir string

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

	rootCmd.Execute()
}

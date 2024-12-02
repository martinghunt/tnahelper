package download

/*
Notes on accessions and genome downloads.

Assemblies:
can be downloaded using 'datasets' REST API. These are accessions
like GCF_000008865.2 or GCA_000008865.2. The URL is:
https://api.ncbi.nlm.nih.gov/datasets/v2/genome/accession/GCF_000008865.2/download?include_annotation_type=GENOME_FASTA&include_annotation_type=GENOME_GFF

It gives you a zip file which is a bit of a zip bomb. Unzipping gives you
README.md and md5sum.txt files in CWD. And then ncbi_dataset/ directory that
has the files we want. Output from unzip:
Archive:  test.zip
  inflating: README.md
  inflating: ncbi_dataset/data/assembly_data_report.jsonl
  inflating: ncbi_dataset/data/GCF_000008865.2/GCF_000008865.2_ASM886v2_genomic.fna
  inflating: ncbi_dataset/data/GCF_000008865.2/genomic.gff
  inflating: ncbi_dataset/data/dataset_catalog.json
  inflating: md5sum.txt

We'll want the *.fna and genomic.gff files.


GenBank:
accessions like NC_000913.3 can be downloaded with 2 calls: one for the fasta
and the other for the gff3:
https://www.ncbi.nlm.nih.gov/sviewer/viewer.fcgi?id=NC_000913.3&db=nuccore&report=gff3&retmode=text
https://www.ncbi.nlm.nih.gov/sviewer/viewer.fcgi?id=NC_000913&db=nuccore&report=fasta&retmode=text

*/

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"github.com/martinghunt/tnahelper/utils"
	"github.com/martinghunt/tnahelper/seqfiles"
)

const BLAST_FTP_URL = "https://ftp.ncbi.nlm.nih.gov/blast/executables/blast+/2.16.0/"

var BLAST_TARBALLS = map[string]string{
	"windows/amd64": "ncbi-blast-2.16.0+-x64-win64.tar.gz",
	//"windows/arm64": "CAN'T GET", no binaries for this
	"darwin/amd64": "ncbi-blast-2.16.0+-x64-macosx.tar.gz",
	"darwin/arm64": "ncbi-blast-2.16.0+-aarch64-macosx.tar.gz",
	"linux/amd64":  "ncbi-blast-2.16.0+-x64-linux.tar.gz",
	"linux/arm64":  "ncbi-blast-2.16.0+-aarch64-linux.tar.gz",
}

func downloadFile(url, outfile string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	fout, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer fout.Close()

	_, err = io.Copy(fout, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func extractGzipTarball(tarball string, outdir string, wantedFiles []string) error {
	file, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)

	wantedMap := make(map[string]struct{})
	for _, file := range wantedFiles {
		wantedMap[file] = struct{}{}
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fmt.Println("Filename:", header.Name)

		if len(wantedFiles) > 0 {
			if _, wanted := wantedMap[filepath.Base(header.Name)]; !wanted {
				continue
			}
		}

		fmt.Println("Extracting:", header.Name)
		outfile := filepath.Join(outdir, filepath.Base(header.Name))

		if err := os.MkdirAll(filepath.Dir(outfile), os.ModePerm); err != nil {
			return err
		}
		fout, err := os.Create(outfile)
		if err != nil {
			return err
		}
		if _, err := io.Copy(fout, tarReader); err != nil {
			fout.Close()
			return err
		}
		fout.Close()
		if runtime.GOOS != "windows" {
			err := os.Chmod(outfile, 0755)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func downloadBlast(outdir string) error {
	key := runtime.GOOS + "/" + runtime.GOARCH
	fmt.Println("Detected OS/architecture:", key)
	tarballToDownload, exists := BLAST_TARBALLS[key]
	if !exists {
		return fmt.Errorf("Cannot download blast binaries. Unknown OS/architecture: %v", key)
	}
	url := BLAST_FTP_URL + tarballToDownload
	tmpOut := filepath.Join(outdir, tarballToDownload)
	fmt.Println("Downloading", url, "to", tmpOut)
	err := downloadFile(url, tmpOut)
	if err != nil {
		return fmt.Errorf("Error downloading %v. Error: %v", url, err)
	}
	// Comment out tblastx for now, since we don't support it but
	// might in the future
	var wanted []string
	if runtime.GOOS == "windows" {
		wanted = []string{"nghttp2.dll", "blastn.exe", "blastn.exe.manifest", "makeblastdb.exe", "makeblastdb.exe.manifest"} //, "tblastx", "tblastx.exe"}
	} else {
		wanted = []string{"blastn", "makeblastdb"} //, "tblastx"}
	}
	extractGzipTarball(tmpOut, outdir, wanted)
	err = os.Remove(tmpOut)
	if err != nil {
		return fmt.Errorf("Error deleting downloaded tarball %v %v", tmpOut, err)
	}
	return nil
}

func DownloadBinaries(outdir string) {
	err := os.MkdirAll(outdir, 0755)
	if err != nil {
		log.Fatalf("Error making output directory %v %v", outdir, err)
	}
	err = downloadBlast(outdir)
	if err != nil {
		log.Fatalf("Error downloading blast binaries %v", err)
	}
}


func FastaAndGffFromZip(zipfile string, outprefix string) {
	zipReader, err := zip.OpenReader(zipfile)
	if err != nil {
		log.Fatalf("Error opening ZIP file %s: %v", zipfile, err)
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		outfile := ""
		if strings.HasSuffix(file.Name, ".fna") {
			outfile = outprefix + ".fa"
			fmt.Println("Extract fasta", file.Name, "to", outfile)
		} else if strings.HasSuffix(file.Name, ".gff") {
			outfile = outprefix + ".gff"
			fmt.Println("Extract gff", file.Name, "to", outfile)
		}
		if outfile == "" {
			continue
		}
		toExtract, err := file.Open()
		if err != nil {
			log.Fatalf("Error opening %s: %v", file.Name, err)
		}
		defer toExtract.Close()

		fout, err := os.Create(outfile)
		if err != nil {
			log.Fatalf("Error opening %s: %v", outfile, err)
		}
		defer fout.Close()

		_, err = io.Copy(fout, toExtract)
		if err != nil {
			log.Fatalf("Error copying contents of file from %s to %s: %v", file.Name, outfile, err)
		}
	}
}


func DownloadGenomeWithDatasetsAPI(accession string, outprefix string, result chan error) {
	tmp_dir := outprefix + ".tmp"
	err := os.MkdirAll(tmp_dir, 0755)
	if err != nil {
		log.Fatalf("Error making temp dir for downloaded files: %s", tmp_dir)
	}
	tmp_zip_file := filepath.Join(tmp_dir, "dl.zip")
	url := fmt.Sprintf("https://api.ncbi.nlm.nih.gov/datasets/v2/genome/accession/%s/download?include_annotation_type=GENOME_FASTA&include_annotation_type=GENOME_GFF", accession)
	err = downloadFile(url, tmp_zip_file)
	if err != nil {
		log.Fatalf("Error downloading '%s' from NCBI (url: %s)\nError: %s", accession, url, err)
	}

	fmt.Println("Downloaded zip file from URL:", url)
	FastaAndGffFromZip(tmp_zip_file, outprefix)
	utils.DeleteFileIfExists(tmp_zip_file)
	utils.DeleteFileIfExists(tmp_dir)
	result <- nil
}


func DownloadGenomeFromGenBank(accession string, outprefix string, result chan error) {
	url := fmt.Sprintf("https://www.ncbi.nlm.nih.gov/sviewer/viewer.fcgi?id=%s&db=nuccore&report=fasta&retmode=text", accession)
	fastaOut := outprefix + ".fa"
	fmt.Println("Getting FASTA: ", url)
	err := downloadFile(url, fastaOut)
	if err != nil {
		log.Fatalf("Error downloading FASTA file for '%s' from NCBI (url: %s)\nError: %s", accession, url, err)
	}

	url = fmt.Sprintf("https://www.ncbi.nlm.nih.gov/sviewer/viewer.fcgi?id=%s&db=nuccore&report=gff3&retmode=text", accession)
	gffOut := outprefix + ".gff"
	fmt.Println("Getting GFF (if it exists): ", url)
	err = downloadFile(url, gffOut)
	if err != nil {
		fmt.Println("Didn't get gff file, but carrying on because FASTA is ok")
	}
	result <- nil
}


func DownloadGenome(accession string, outprefix string) error {
	downloadErr := make(chan error)

	if strings.HasPrefix(accession, "GCF_") || strings.HasPrefix(accession, "GCA_") {
		go DownloadGenomeWithDatasetsAPI(accession, outprefix, downloadErr)
	} else {
		go DownloadGenomeFromGenBank(accession, outprefix, downloadErr)
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
loop:
	for {
		select {
		case err := <-downloadErr:
			if err != nil {
				log.Fatalf("Error downloading %s", accession)
			}
			fmt.Println("Finished downloading", accession)
            break loop
		case <- ticker.C:
			fmt.Print(".")
		}
	}
    dlFa := outprefix + ".fa"
    tmpFa := outprefix + ".tmp.fa"
    seqfiles.ParseSeqFile(dlFa, outprefix + ".tmp")
    utils.RenameFile(tmpFa, dlFa)
    return nil
}


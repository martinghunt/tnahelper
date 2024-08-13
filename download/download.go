package download

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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
		return fmt.Errorf("Cannot download blast binaries. Unknown OS/architecture:", key)
	}
	url := BLAST_FTP_URL + tarballToDownload
	tmpOut := filepath.Join(outdir, tarballToDownload)
	fmt.Println("Downloading", url, "to", tmpOut)
	err := downloadFile(url, tmpOut)
	if err != nil {
		return fmt.Errorf("Error downloading", url, " ... error:", err)
	}
	var wanted []string
	if runtime.GOOS == "windows" {
		wanted = []string{"blastn.exe", "blastn.exe.manifest", "makeblastdb.exe", "makeblastdb.exe.manifest"}
	} else {
		wanted = []string{"blastn", "makeblastdb"}
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

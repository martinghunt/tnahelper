package blast

import (
	"fmt"
	"github.com/martinghunt/tnahelper/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
    "runtime"
)

func RunBlast(workingDir string, binDir string) {
	makeblastdb := filepath.Join(binDir, "makeblastdb")
	blastn := filepath.Join(binDir, "blastn")
	if runtime.GOOS == "windows" {
        makeblastdb += ".exe"
        blastn += ".exe"
    }

	tempDir, err := os.MkdirTemp("", "tna-blast-")
	if err != nil {
		log.Fatalf("Failed to create temporary dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	fmt.Println("Working in temporary dir:", tempDir)

	refToCopy := filepath.Join(workingDir, "g2.fa")
	ref := filepath.Join(tempDir, "ref.fa")
	utils.CopyFile(refToCopy, ref)
	blastdb := filepath.Join(tempDir, "blast_db")
	fmt.Println("Running makeblastdb", makeblastdb)
	command := exec.Command(makeblastdb, "-dbtype", "nucl", "-in", ref, "-out", blastdb)
	output, err := command.CombinedOutput()
	if err != nil {
		os.RemoveAll(tempDir)
		log.Fatalf("Error running makeblastdb: %s\n%s", output, err)
	}
	fmt.Printf("output: %s", output)

	fmt.Println("Running blastn", blastn)
	blastn_out_tmp := filepath.Join(tempDir, "blast_db")
	blastn_out := filepath.Join(workingDir, "blast")
	qryToCopy := filepath.Join(workingDir, "g1.fa")
	qry := filepath.Join(tempDir, "qry.fa")
	utils.CopyFile(qryToCopy, qry)
	command = exec.Command(blastn, "-db", blastdb, "-query", qry, "-out", blastn_out_tmp, "-outfmt", "6 qseqid sseqid pident qstart qend sstart send qseq sseq")
	output, err = command.CombinedOutput()
	if err != nil {
		os.RemoveAll(tempDir)
		log.Fatalf("Error running blastn: %s\n%s", output, err)
	}
	fmt.Println("Finished running blastn")
	utils.CopyFile(blastn_out_tmp, blastn_out)
	fmt.Println("Tidied up temproary files")
}

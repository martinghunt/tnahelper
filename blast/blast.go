package blast

import (
	"fmt"
	"github.com/martinghunt/tnahelper/utils"
	"github.com/shenwei356/xopen"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type AlnBlock struct {
	qstart  int
	qend    int
	rstart  int
	rend    int
	alnType int
}

func ParseBlastFile(infile string, outfile string, blastType string) {
	reader, err := xopen.Ropen(infile)
	if err != nil {
		log.Fatalf("Error opening blast file %v: %v", infile, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Fatalf("Error closing file %v: %v", infile, err)
		}
	}()
	fout, errOut := xopen.Wopen(outfile)
	if errOut != nil {
		log.Fatalf("Error opening blast file for writing %v: %v", outfile, errOut)
	}
	defer fout.Close()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf("read file line error: %v", err)
			return
		}

		fields := strings.Split(strings.TrimSpace(line), "\t")
		// fields are:
		// 0, 1 = ref, qry name
		// 2 = percent identity
		// 3, 4 = ref start/end
		// 5, 6 = qry start/end
		// 7, 8 = ref/qry alignment string

		// tblastx can have query start > query end. blastn does not.
		// TNA needs query start < end. So if needed, swap the start/end
		// coordinates around and reverse the alignment strings
		var qstart, _ = strconv.Atoi(fields[3])
		var qend, _ = strconv.Atoi(fields[4])
		if qstart > qend {
			if blastType == "blastn" {
				log.Fatalf("Query start > end, and using blastn. Cannot continue\n%v", fields)
			}
			fields[4], fields[3] = fields[3], fields[4]
			fields[6], fields[5] = fields[5], fields[6]
			fields[7] = utils.Reverse(fields[7])
			fields[8] = utils.Reverse(fields[8])
		}

		fout.WriteString(strings.Join(fields[:7], "\t") + "\t[")
		var rpos = 0
		var qpos = 0
		var alnBlocks = []AlnBlock{
			{qstart: 0, qend: 0, rstart: 0, rend: 0, alnType: 0},
		}

		for i := 1; i < len(fields[7]); i++ {
			if fields[7][i] == '-' {
				if fields[8][i] == '-' {
					log.Fatalf("Error, both seqs have gap at same position. Cannot continue\n%v\n%v", fields[7], fields[8])
				}
				rpos++
			} else if fields[8][i] == '-' {
				qpos++
			} else if fields[7][i] == fields[8][i] {
				if alnBlocks[len(alnBlocks)-1].qend == qpos && alnBlocks[len(alnBlocks)-1].rend == rpos && alnBlocks[len(alnBlocks)-1].alnType == 0 {
					alnBlocks[len(alnBlocks)-1].qend++
					alnBlocks[len(alnBlocks)-1].rend++
				} else {
					alnBlocks = append(alnBlocks, AlnBlock{qstart: qpos + 1, qend: qpos + 1, rstart: rpos + 1, rend: rpos + 1, alnType: 0})
				}
				rpos++
				qpos++
			} else {
				rpos++
				qpos++
				alnBlocks = append(alnBlocks, AlnBlock{qstart: qpos, qend: qpos, rstart: rpos, rend: rpos, alnType: 1})
			}
		}
		for i, a := range alnBlocks {
			if i > 0 {
				fout.WriteString(",")
			}
			if blastType == "blastn" {
				fmt.Fprintf(fout, "[%d,%d,%d,%d,%d]", a.qstart, a.qend, a.rstart, a.rend, a.alnType)
			} else if a.alnType == 0 {
				fmt.Fprintf(fout, "[%d,%d,%d,%d,%d]", 3*a.qstart, 3*a.qend+2, 3*a.rstart, 3*a.rend+2, a.alnType)
			} else {
				for j := 0; j < 3; j++ {
					fmt.Fprintf(fout, "[%d,%d,%d,%d,%d]", j+3*a.qstart, j+3*a.qend, j+3*a.rstart, j+3*a.rend, a.alnType)
					if j < 2 {
						fout.WriteString(",")
					}
				}
			}
		}

		fout.WriteString("]\n")
	}
}

func RunBlast(workingDir string, binDir string, blastType string, extraOptions []string) {
	fmt.Println("Extra options:", extraOptions)
	makeblastdb := filepath.Join(binDir, "makeblastdb")
	blastProgram := filepath.Join(binDir, blastType)
	if runtime.GOOS == "windows" {
		makeblastdb += ".exe"
		blastProgram += ".exe"
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

	fmt.Println("Running blast", blastProgram)
	blast_out_tmp := filepath.Join(tempDir, "blast_db")
	blast_out := filepath.Join(workingDir, "blast")
	qryToCopy := filepath.Join(workingDir, "g1.fa")
	qry := filepath.Join(tempDir, "qry.fa")
	utils.CopyFile(qryToCopy, qry)
	var commandline = append([]string{"-db", blastdb, "-query", qry, "-out", blast_out_tmp, "-outfmt", "6 qseqid sseqid pident qstart qend sstart send qseq sseq"}, extraOptions...)
	fmt.Println("Going to run this blast command:", blastProgram, strings.Join(commandline, " "))
	command = exec.Command(blastProgram, commandline...)
	output, err = command.CombinedOutput()
	if err != nil {
		os.RemoveAll(tempDir)
		log.Fatalf("Error running blast: %s\n%s", output, err)
	}
	fmt.Println("Finished running blast")
	ParseBlastFile(blast_out_tmp, blast_out, blastType)
	fmt.Println("Tidied up temproary files")
}

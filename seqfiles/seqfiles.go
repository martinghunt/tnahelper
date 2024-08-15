package seqfiles

import (
	"fmt"
	"github.com/shenwei356/xopen"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type FileFormat uint64

const (
	Unknown FileFormat = iota
	FASTA
	FASTQ
	GFF3
	GENBANK
)

func getFileType(filename string) FileFormat {
	reader, err := xopen.Ropen(filename)
	if err != nil {
		log.Fatalf("Error opening file %v: %v", filename, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Fatalf("Error closing file %v: %v", filename, err)
		}
	}()

	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading fiest line of file %v: %v", filename, err)
	}

	if strings.HasPrefix(line, ">") {
		return FASTA
	} else if strings.HasPrefix(line, "@") {
		return FASTQ
	} else if strings.HasPrefix(line, "##gff-version") {
		return GFF3
	} else if strings.HasPrefix(line, "LOCUS ") {
		return GENBANK
	}
	log.Fatalf("Cannot determine type of input file %v", filename)
	return Unknown
}

func parseFastaFile(infile string, outfile string) {
	reader, err := xopen.Ropen(infile)
	if err != nil {
		log.Fatalf("Error opening file %v: %v", infile, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Fatalf("Error closing file %v: %v", infile, err)
		}
	}()
	fout, errOut := xopen.Wopen(outfile)
	if errOut != nil {
		log.Fatalf("Error opening file for writing %v: %v", outfile, errOut)
	}
	defer fout.Close()
	first := true

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf("read file line error: %v", err)
			return
		}

		if strings.HasPrefix(line, ">") {
			if first {
				first = false
			} else {
				fout.WriteString("\n")
			}
			fout.WriteString(strings.Fields(line)[0] + "\n")
		} else {
			fout.WriteString(strings.ToUpper(strings.TrimSpace(line)))
		}
	}
	fout.WriteString("\n")
	fout.Flush()
}

func parseFastqFile(infile string, outfile string) {
	reader, err := xopen.Ropen(infile)
	if err != nil {
		log.Fatalf("Error opening file %v: %v", infile, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Fatalf("Error closing file %v: %v", infile, err)
		}
	}()
	fout, errOut := xopen.Wopen(outfile)
	if errOut != nil {
		log.Fatalf("Error opening file for writing %v: %v", outfile, errOut)
	}
	defer fout.Close()
	oneRead := [4]string{}
	lastRead := false

	for {
		for i := 0; i < 4; i++ {
			oneRead[i], err = reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					lastRead = true
					break
				}

				log.Fatalf("read file line error: %v", err)
				return
			}
		}
		if lastRead {
			break
		}

		if !(strings.HasPrefix(oneRead[0], "@") && strings.HasPrefix(oneRead[2], "+")) {
			log.Fatalf("Error getting sequence from file %v, around here: %v%v", infile, oneRead[0], oneRead[1])
		}

		fout.WriteString(">")
		fout.WriteString(oneRead[0][1:])
		fout.WriteString(strings.ToUpper(oneRead[1]))
	}
	fout.Flush()
}

func parseGenbankFile(infile string, outfileSeqs string, outfileAnnot string) {
	reader, err := xopen.Ropen(infile)
	if err != nil {
		log.Fatalf("Error opening file %v: %v", infile, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Fatalf("Error closing file %v: %v", infile, err)
		}
	}()
	foutSeqs, errOut := xopen.Wopen(outfileSeqs)
	if errOut != nil {
		log.Fatalf("Error opening sequence file for writing %v: %v", outfileSeqs, errOut)
	}
	foutAnnot, errOut := xopen.Wopen(outfileAnnot)
	if errOut != nil {
		log.Fatalf("Error opening annotation file for writing %v: %v", outfileAnnot, errOut)
	}
	defer foutSeqs.Close()
	defer foutAnnot.Close()
	inGene := false
	inSeq := false
	inFeatures := false
	inHeader := false
	geneStart := 0
	geneEnd := 0
	geneStrand := ""
	geneName := ""
	seqReplaceRe := regexp.MustCompile(`[\s0-9]`)
	coordsReplaceRe := regexp.MustCompile(`[^0-9\.]`)
	genePrefix := "                     /gene="
	locustagPrefix := "                     /locus_tag"
	currentContig := "UNKNOWN"

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf("read file line error: %v", err)
			return
		}

		if line == "//\n" {
			foutSeqs.WriteString("\n")
			inSeq = false
			inGene = false
		} else if inSeq {
			foutSeqs.WriteString(strings.ToUpper(seqReplaceRe.ReplaceAllString(line, "")))
		} else if inHeader {
			if strings.HasPrefix(line, "FEATURES") {
				inHeader = false
				inFeatures = true
			}
		} else if inFeatures {
			if strings.HasPrefix(line, "ORIGIN") {
				inFeatures = false
				if inGene {
					foutAnnot.WriteString(fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", currentContig, geneName, geneStart, geneEnd, geneStrand))
					geneName = ""
					geneStart = 0
					geneEnd = 0
					geneStrand = "."
					inGene = false
				}
				inSeq = true
			}

			if inGene {
				if strings.HasPrefix(line, genePrefix) || strings.HasPrefix(line, locustagPrefix) {
					fields := strings.Split(line, "=")
					if geneName == "UNKNOWN" {
						geneName = ""
					} else if len(geneName) > 0 {
						geneName += "/"
					}
					geneName += strings.TrimSpace(strings.ReplaceAll(fields[1], "\"", ""))
				} else if line[5] != ' ' {
					foutAnnot.WriteString(fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", currentContig, geneName, geneStart, geneEnd, geneStrand))
					inGene = false
					geneName = ""
					geneStart = 0
					geneEnd = 0
					geneStrand = "."
				}

			}

			if strings.HasPrefix(line, "     gene") {
				inGene = true
				fields := strings.Fields(line)
				if strings.HasPrefix(fields[1], "complement") {
					geneStrand = "-"
				} else {
					geneStrand = "+"
				}

				fields[1] = coordsReplaceRe.ReplaceAllString(fields[1], "")
				coords := strings.Split(fields[1], "..")
				geneStart, _ = strconv.Atoi(coords[0])
				geneEnd, _ = strconv.Atoi(coords[1])
				if geneEnd < geneStart {
					geneStart, geneEnd = geneEnd, geneStart
				}
				geneName = "UNKNOWN"
			}
			continue
		} else if strings.HasPrefix(line, "LOCUS ") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				log.Fatalf("Error getting sequence name from LOCUS line of genbank file %v: %v", infile, line)
			}
			foutSeqs.WriteString(">")
			currentContig = fields[1]
			foutSeqs.WriteString(currentContig)
			foutSeqs.WriteString("\n")
			inHeader = true
		}
	}
	foutSeqs.Flush()
	foutAnnot.Flush()
}

func nameFromGff3Col8(col8 string) string {
	keysVals := strings.Split(col8, ";")
	names := []string{}
	for _, s := range keysVals {
		if strings.HasPrefix(s, "ID=") || strings.HasPrefix(s, "name=") {
			names = append(names, strings.TrimSpace(strings.Split(s, "=")[1]))
		}
	}
	return strings.Join(names, "/")
}

func parseGFF3File(infile string, outfileSeqs string, outfileAnnot string) {
	reader, err := xopen.Ropen(infile)
	if err != nil {
		log.Fatalf("Error opening file %v: %v", infile, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Fatalf("Error closing file %v: %v", infile, err)
		}
	}()
	foutSeqs, errOut := xopen.Wopen(outfileSeqs)
	if errOut != nil {
		log.Fatalf("Error opening sequence file for writing %v: %v", outfileSeqs, errOut)
	}
	foutAnnot, errOut := xopen.Wopen(outfileAnnot)
	if errOut != nil {
		log.Fatalf("Error opening annotation file for writing %v: %v", outfileAnnot, errOut)
	}
	defer foutSeqs.Close()
	defer foutAnnot.Close()
	firstSeq := true
	inFasta := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf("read file line error: %v", err)
			return
		}

		if strings.HasPrefix(line, "##FASTA") {
			inFasta = true
			continue
		} else if line[0] == '#' {
			continue
		}

		if inFasta {
			if strings.HasPrefix(line, ">") {
				if firstSeq {
					firstSeq = false
				} else {
					foutSeqs.WriteString("\n")
				}
				foutSeqs.WriteString(line)
			} else {
				foutSeqs.WriteString(strings.ToUpper(strings.TrimSpace(line)))
			}
		} else {
			fields := strings.Split(line, "\t")
			if fields[2] == "gene" {
				name := nameFromGff3Col8(fields[8])
				foutAnnot.WriteString(fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", fields[0], name, fields[3], fields[4], fields[6]))
			}
		}
	}
	foutSeqs.WriteString("\n")
	foutSeqs.Flush()
	foutAnnot.Flush()
}

func ParseSeqFile(infile string, outprefix string) {
	filetype := getFileType(infile)
	fastaOutfile := outprefix + ".fa"
	annotOutfile := outprefix + ".annot"
	switch filetype {
	case FASTA:
		parseFastaFile(infile, fastaOutfile)
	case FASTQ:
		parseFastqFile(infile, fastaOutfile)
	case GFF3:
		parseGFF3File(infile, fastaOutfile, annotOutfile)
	case GENBANK:
		parseGenbankFile(infile, fastaOutfile, annotOutfile)
	default:
		log.Fatalf("Error: could not determine type of file %v", infile)
	}
}

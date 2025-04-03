package seqfiles

import (
	"fmt"
	"github.com/martinghunt/tnahelper/utils"
	"github.com/shenwei356/xopen"
	"io"
	"log"
	"os"
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
	EMBL
)

type Gap struct {
	SeqName string
	Start   int
	End     int
}

func GetFileType(filename string) FileFormat {
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
	} else if strings.HasPrefix(line, "ID ") {
		return EMBL
	}
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
				if len(line) > 0 { // ie last line of file has no newline
					fout.WriteString(strings.ToUpper(strings.TrimSpace(line)))
				}
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

func seqnameFromLineGenbankOrEMBL(line string, fformat FileFormat) string {
	if fformat == GENBANK && strings.HasPrefix(line, "LOCUS ") {
		fields := strings.Fields(strings.TrimRight(line, "\n"))
		if len(fields) < 2 {
			log.Fatalf("Error getting sequence name from LOCUS line of genbank file: %v", line)
		}
		return fields[1]
	} else if fformat == EMBL && strings.HasPrefix(line, "ID ") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			log.Fatalf("Error getting sequence name from LOCUS line of EMBL file: %v", line)
		}
		return strings.TrimRight(fields[1], ";")
	}
	return ""
}

func endGenbankOrEmblHeader(line string, fformat FileFormat) bool {
	if fformat == GENBANK {
		return strings.HasPrefix(line, "FEATURES")
	} else if fformat == EMBL {
		return line == "FH\n"
	} else {
		log.Fatalf("Unknown file format: %v", fformat)
	}
	panic("Unexpectedly reached an invalid state in endGenbankOrEmblHeader")
}

func lineMarksGebnkaOrEmblSequenceStart(line string, fformat FileFormat) bool {
	if fformat == GENBANK {
		return strings.HasPrefix(line, "ORIGIN")
	} else if fformat == EMBL {
		return strings.HasPrefix(line, "SQ   ")
	} else {
		log.Fatalf("Uknown file format: %v", fformat)
	}
	panic("Unexpectedly reached an invalid state in lineMarksGebnkaOrEmblSequenceStart")
}

func parseGenbankOrEmblFile(infile string, outfileSeqs string, outfileAnnot string, fformat FileFormat) {
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
	genePrefix := "                     /gene="
	locustagPrefix := "                     /locus_tag"
	currentContig := "UNKNOWN"
	foutAnnot.WriteString("##gff-version 3\n")
	nonNumberRe := regexp.MustCompile(`\D+`)

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
			if endGenbankOrEmblHeader(line, fformat) {
				inHeader = false
				inFeatures = true
			}
		} else if inFeatures {
			// Genbank and EMBL feature lines are the same, except that EMBL
			// startswith "FT", whereas Genbank is spaces
			if fformat == EMBL && strings.HasPrefix(line, "FT ") {
				line = strings.Replace(line, "FT", "  ", 1)
			}

			if lineMarksGebnkaOrEmblSequenceStart(line, fformat) {
				inFeatures = false
				if inGene {
					foutAnnot.WriteString(fmt.Sprintf("%v\t.\tgene\t%v\t%v\t.\t%v\t.\tID=%v\n", currentContig, geneStart, geneEnd, geneStrand, geneName))
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
					foutAnnot.WriteString(fmt.Sprintf("%v\t.\tgene\t%v\t%v\t.\t%v\t.\tID=%v\n", currentContig, geneStart, geneEnd, geneStrand, geneName))
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

				coords := nonNumberRe.Split(fields[1], -1)
				geneStart = -1
				geneEnd = -1
				for _, s := range coords {
					if s == "" {
						continue
					}
					c, _ := strconv.Atoi(s)
					if geneStart == -1 || c < geneStart {
						geneStart = c
					}
					if geneEnd == -1 || c > geneEnd {
						geneEnd = c
					}
				}
				geneName = "UNKNOWN"
			}
			continue
		} else {
			seqname := seqnameFromLineGenbankOrEMBL(line, fformat)
			if seqname == "" {
				continue
			}
			foutSeqs.WriteString(">")
			currentContig = seqname
			foutSeqs.WriteString(currentContig)
			foutSeqs.WriteString("\n")
			inHeader = true
		}
	}
	foutSeqs.Flush()
	foutAnnot.Flush()
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
	foutAnnot.WriteString("##gff-version 3\n")

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
			foutAnnot.WriteString(line)
		}
	}
	foutSeqs.WriteString("\n")
	foutSeqs.Flush()
	foutAnnot.Flush()
}

func getGapsFromSingleLineFasta(infile string, minimumGapLen ...int) []Gap {
	minGapLen := 1
	if len(minimumGapLen) > 0 {
		minGapLen = minimumGapLen[0]
	}
	gaps := []Gap{}
	if minGapLen <= 0 {
		return gaps
	}
	gapRegex := regexp.MustCompile(fmt.Sprintf(`N{%d,}`, minGapLen))

	reader, err := xopen.Ropen(infile)
	if err != nil {
		log.Fatalf("Error opening file %v: %v", infile, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Fatalf("Error closing file %v: %v", infile, err)
		}
	}()
	currentName := ""

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf("read file line error: %v", err)
		}

		if strings.HasPrefix(line, ">") {
			fields := strings.Fields(line)
			currentName = strings.TrimPrefix(fields[0], ">")
		} else {
			for _, match := range gapRegex.FindAllStringIndex(line, -1) {
				gaps = append(gaps, Gap{SeqName: currentName, Start: match[0] + 1, End: match[1]})
			}
		}
	}
	return gaps
}

func addGapsToAnnotFile(gaps []Gap, filename string) {
	annotFileExists := utils.FileExists(filename)
	fout, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer fout.Close()
	if !(annotFileExists) {
		fout.WriteString("##gff-version 3\n")
	}
	for _, gap := range gaps {
		fout.WriteString(fmt.Sprintf("%v\tTNA\tgap\t%v\t%v\t.\t+\t.\tname=gap\n", gap.SeqName, gap.Start, gap.End))
	}
}

func ParseSeqFile(infile string, outprefix string, minimumGapLen ...int) {
	minGapLen := 1
	if len(minimumGapLen) > 0 {
		minGapLen = minimumGapLen[0]
	}
	filetype := GetFileType(infile)
	fastaOutfile := outprefix + ".fa"
	annotOutfile := outprefix + ".gff"
	switch filetype {
	case FASTA:
		parseFastaFile(infile, fastaOutfile)
	case FASTQ:
		parseFastqFile(infile, fastaOutfile)
	case GFF3:
		parseGFF3File(infile, fastaOutfile, annotOutfile)
	case GENBANK:
		parseGenbankOrEmblFile(infile, fastaOutfile, annotOutfile, GENBANK)
	case EMBL:
		parseGenbankOrEmblFile(infile, fastaOutfile, annotOutfile, EMBL)
	default:
		log.Fatalf("Error: could not determine type of file %v", infile)
	}
	gaps := getGapsFromSingleLineFasta(fastaOutfile, minGapLen)
	if len(gaps) > 0 {
		addGapsToAnnotFile(gaps, annotOutfile)
	}
}

package example_data

import (
	"github.com/martinghunt/tnahelper/utils"
	"github.com/shenwei356/xopen"
	"log"
	"math/rand"
	"path/filepath"
)

func init() {
	rand.Seed(42)
}

func makeRandomSeq(length int) []byte {
	acgt := []byte{'A', 'C', 'G', 'T'}
	seq := make([]byte, length)
	for i := 0; i < length; i++ {
		seq[i] = acgt[rand.Intn(len(acgt))]
	}
	return seq
}

func mutateNBases(seq []byte, numSnps int) []byte {
	mutated := make([]byte, len(seq))
	copy(mutated, seq)
	step := len(seq) / numSnps

	for i := 0; i < numSnps; i++ {
		pos := step * i
		switch mutated[pos] {
		case 'A':
			mutated[pos] = 'C'
		case 'C':
			mutated[pos] = 'G'
		case 'G':
			mutated[pos] = 'T'
		case 'T':
			mutated[pos] = 'A'
		}
	}

	return mutated
}

/*
Make two genomes, plus matches, that make it look like "TNA" when viewed
with TNA. Contigs like:

<------- c1 -------> <c2> <c3> <-c4-> <c5> <-- c6 --><c7>
       |    |             |  |\       |  |          / /\ \
       |    |             |  |  \     |  |         / /  \ \
       |    |             |  |     \  |  |        / /    \ \
<-c1-> <-c2-> <--c3 ----> <c4> <-c5-> <c6> <c7> <c8><c9><c10>

Bad ascii art, but need top c3 to match bottom c6.
But don't want bottom c4 to match top c5. Which means need to add mutations
to make them too different for blast
*/

func MakeTestFiles(outdir string) {
	commonT := makeRandomSeq(500)
	bot4 := makeRandomSeq(500)
	top3 := mutateNBases(bot4, 13)
	bot6 := mutateNBases(top3, 37)
	top5 := mutateNBases(bot6, 17)
	top7 := makeRandomSeq(500)
	bot8 := mutateNBases(top7, 10)
	bot10 := mutateNBases(top7, 21)

	genome1_gff := filepath.Join(outdir, "g1.gff")
	fout, errOut := xopen.Wopen(genome1_gff)
	if errOut != nil {
		log.Fatalf("Error opening file for writing %v: %v", genome1_gff, errOut)
	}
	defer fout.Close()
	fout.WriteString("##gff-version 3\n")
	fout.WriteString("g1.c1\t.\tgene\t900\t1400\t.\t+\t.\tID=gene1;foo=bar;name=name1\n")
	fout.WriteString("g1.c1\t.\texon\t910\t1000\t.\t+\t.\tID=exon1;name=gene1.exon1;Parent=gene1\n")
	fout.WriteString("g1.c1\t.\texon\t1100\t1400\t.\t+\t.\tID=exon2;name=gene1.exon2;Parent=gene1\n")
	fout.WriteString("g1.c3\t.\tgene\t100\t400\t.\t-\t.\tID=gene2;name=name2\n")
	fout.WriteString("g1.c7\t.\tfeature\t50\t300\t.\t+\t.\tID=feature1;name=feature1_name;color=blue\n")
	fout.WriteString("g1.c7\t.\tfeature\t290\t470\t.\t-\t.\tID=feature2;name=feature2_name\n")
	fout.WriteString("##FASTA\n")
	fout.WriteString(">g1.c0\n")
	fout.WriteString(string(makeRandomSeq(100)) + "\n")
	fout.WriteString(">g1.c1\n")
	fout.WriteString(string(makeRandomSeq(900)))
	fout.WriteString(string(commonT))
	fout.WriteString(string(makeRandomSeq(900)) + "\n")
	fout.WriteString(">g1.c2\n")
	fout.WriteString(string(makeRandomSeq(100)) + "\n")
	fout.WriteString(">g1.c3\n")
	fout.WriteString(string(top3) + "\n")
	fout.WriteString(">g1.c4\n")
	fout.WriteString(string(makeRandomSeq(800)) + "\n")
	fout.WriteString(">g1.c5\n")
	fout.WriteString(string(utils.ReverseComplement(top5)) + "\n")
	fout.WriteString(">g1.c6\n")
	fout.WriteString(string(makeRandomSeq(800)) + "\n")
	fout.WriteString(">g1.c7\n")
	fout.WriteString(string(top7) + "\n")
	fout.Flush()
	fout.Close()

	genome2_gff := filepath.Join(outdir, "g2.gff")
	fout, errOut = xopen.Wopen(genome2_gff)
	if errOut != nil {
		log.Fatalf("Error opening file for writing %v: %v", genome2_gff, errOut)
	}
	defer fout.Close()
	fout.WriteString("##gff-version 3\n")
	fout.WriteString("g2.c2\t.\tgene\t1\t500\t.\t-\t.\tID=gene42;foo=bar;name=name_gene42\n")
	fout.WriteString("g2.c10\t.\tgene\t150\t410\t.\t+\t.\tID=gene43;foo=bar;name=name_gene43\n")
	fout.WriteString("##FASTA\n")
	fout.WriteString(">g2.c1\n")
	fout.WriteString(string(makeRandomSeq(1000)) + "\n")
	fout.WriteString(">g2.c2\n")
	fout.WriteString(string(commonT) + "\n")
	fout.WriteString(">g2.c3\n")
	fout.WriteString(string(makeRandomSeq(1000)) + "\n")
	fout.WriteString(">g2.c4\n")
	fout.WriteString(string(bot4) + "\n")
	fout.WriteString(">g2.c5\n")
	fout.WriteString(string(makeRandomSeq(800)) + "\n")
	fout.WriteString(">g2.c6\n")
	fout.WriteString(string(utils.ReverseComplement(bot6)) + "\n")
	fout.WriteString(">g2.c7\n")
	fout.WriteString(string(makeRandomSeq(400)) + "\n")
	fout.WriteString(">g2.c8\n")
	fout.WriteString(string(bot8) + "\n")
	fout.WriteString(">g2.c9\n")
	fout.WriteString(string(makeRandomSeq(600)) + "\n")
	fout.WriteString(">g2.c10\n")
	fout.WriteString(string(bot10) + "\n")
}

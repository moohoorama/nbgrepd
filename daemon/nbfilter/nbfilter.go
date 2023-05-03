package nbfilter

import (
	"fmt"
	"github.com/cespare/xxhash/v2"
	"github.com/datadog/hyperloglog"
	"log"
	"math"
	"strings"
)

const formatWidth = 16
const formatUnit = 4

type NBFilter struct {
	bits    []byte
	ngram   int
	numhash int

	cardinality uint64
}

func (nbf NBFilter) String() string {
	return fmt.Sprintf("NBFilter(size:%d,ngram:%d,numhash:%d,cardinality:%d)",
		len(nbf.bits)*8, nbf.ngram, nbf.numhash, nbf.cardinality)
}
func (nbf NBFilter) Detail() string {
	sb := strings.Builder{}
	sb.WriteString(nbf.String())
	sb.WriteString(":\n")
	for i := 0; i < len(nbf.bits); i += formatWidth {
		sb.WriteString(fmt.Sprintf("%8d |", i))
		for j := i; j < len(nbf.bits) && j < i+formatWidth; j += formatUnit {
			cur := nbf.bits[j:]
			if len(cur) > formatUnit {
				cur = cur[:formatUnit]
			}
			sb.WriteString(fmt.Sprintf(" %x", []byte(cur)))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (nbf NBFilter) CheckString(word string) bool {
	return nbf.Check([]byte(word))
}

func (nbf NBFilter) Check(word []byte) bool {
	if len(word) < nbf.ngram {
		log.Fatalf("InvalidArgument(len(%x)<%d)", word, nbf.ngram)
	}
	/* ex) ngram=3, word=abcdef
	 *
	 * check(abc)
	 * check(bcd)
	 * check(cde)
	 * check(def) */
	for i := 0; i < len(word)-nbf.ngram; i++ {
		if !nbf.check(word[i : i+nbf.ngram]) {
			return false
		}
	}

	return true
}
func (nbf NBFilter) check(word []byte) bool {
	if len(word) > nbf.ngram {
		log.Fatalf("InvalidArgument(len(%x)>%d)", word, nbf.ngram)
	}
	buf := make([]byte, nbf.ngram+1)
	copy(buf, word)
	for k := 0; k < nbf.numhash; k++ {
		buf[nbf.ngram] = byte(k)
		digest := xxhash.Sum64(buf) % uint64(len(nbf.bits)*8)

		off := digest / 8
		bit := digest % 8

		if (nbf.bits[off]>>bit)&1 != 1 {
			return false
		}
	}
	return true
}

func Create(org []byte, ngram, numhash, filterBytes int) NBFilter {
	buf := make([]byte, ngram+1)

	hllSize := 1 << (int(math.Log2(float64(len(org)))))
	if hllSize < 64 {
		hllSize = 64
	}
	hll, err := hyperloglog.New(uint(hllSize))
	if err != nil {
		log.Fatalf("hyperloglog.New=>%v", err)
	}

	for i := range org {
		word := org[i:]
		if len(word) > ngram {
			word = word[:ngram]
		}
		if len(word) < ngram {
			break
		}
		hll.Add(uint32(xxhash.Sum64(word)))
	}

	filterBytes = int(hll.Count())
	if filterBytes < 64 {
		filterBytes = 64
	}

	bits := make([]byte, filterBytes)
	for i := range org {
		word := org[i:]
		if len(word) > ngram {
			word = word[:ngram]
		}
		if len(word) < ngram {
			break
		}
		copy(buf, word)
		for k := 0; k < numhash; k++ {
			buf[ngram] = byte(k)
			digest := xxhash.Sum64(buf) % uint64(filterBytes*8)

			off := digest / 8
			bit := digest % 8

			bits[off] |= 1 << bit
		}
	}
	return NBFilter{bits, ngram, numhash, hll.Count()}
}

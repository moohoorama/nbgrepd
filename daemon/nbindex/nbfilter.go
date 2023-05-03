package nbindex

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/datadog/hyperloglog"
	"log"
	"math"
	"strings"
)

const formatWidth = 16
const formatUnit = 4

const tailSize = (8 /*uint64*/ * 4 /*ngram,skip,numhash,cardinality*/)

/* NGram Bloom filter로 grep할 대상키가 있는지 없는지를 표시하는 index */
type NBIndex struct {
	/* 원래는 BloomIndex의 bit flag 저장소인데,
	 * marshaling & unmarshaling을 편하게 하기 위해
	 * ngram,skip,numhash,cardinality등을 tail로 붙여놓는다 */
	bits []byte

	ngram       int
	skip        int
	numhash     int
	cardinality uint64
}

func (nbf NBIndex) String() string {
	return fmt.Sprintf("NBIndex(size:%d,ngram:%d,skip:%d,numhash:%d,cardinality:%d)",
		(len(nbf.bits)-tailSize)*8, nbf.ngram, nbf.skip, nbf.numhash, nbf.cardinality)
}
func (nbf NBIndex) Detail() string {
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

func (nbf NBIndex) CheckString(word string) bool {
	return nbf.Check([]byte(word))
}

func (nbf NBIndex) GetBody() []byte {
	return nbf.bits
}

func (nbf NBIndex) Check(word []byte) bool {
	if len(word) < nbf.ngram+nbf.skip {
		log.Fatalf("InvalidArgument(len(%s)<%d+%d)", word, nbf.ngram, nbf.skip)
	}
	/* ex) ngram=4, skip=2, word=abcdef
	 * (check(abcd) and check(cdef)) or (check(bcde)) */
	for s := 0; s < nbf.skip; s++ {
		check := true
		for i := s; i < len(word)-nbf.ngram; i += nbf.skip {
			if !nbf.check(word[i : i+nbf.ngram]) {
				check = false
				break
			}
		}
		if check {
			return true
		}
	}

	return false
}
func (nbf NBIndex) check(word []byte) bool {
	if len(word) > nbf.ngram {
		log.Fatalf("InvalidArgument(len(%x)>%d)", word, nbf.ngram)

	}
	buf := make([]byte, nbf.ngram+1)
	copy(buf, word)
	for k := 0; k < nbf.numhash; k++ {
		buf[nbf.ngram] = byte(k)
		digest := Digest(buf) % uint32((len(nbf.bits)-tailSize)*8)

		off := digest / 8
		bit := digest % 8

		if (nbf.bits[off]>>bit)&1 != 1 {
			return false
		}
	}
	return true
}

func Create(org []byte, ngram, skip, numhash, filterBytes, maxCardinality int) (processedSize int64, nbf *NBIndex, err error) {
	buf := make([]byte, ngram+1)

	if len(org) <= ngram+skip {
		return 0, nil, nil
	}

	if skip < 1 {
		return 0, nil, fmt.Errorf("InvalidArgument(skip:%d)", skip)
	}

	/* cardinality 계산용 */
	hllSize := 1 << (int(math.Log2(float64(len(org)))))
	if hllSize < 64 {
		hllSize = 64
	}
	hllSize = 8192
	hll, err := hyperloglog.New(uint(hllSize))
	if err != nil {
		return 0, nil, fmt.Errorf("hyperloglog.New=>%v", err)
	}

	bits := make([]byte, filterBytes+tailSize)
	/* tail 붙임 */
	binary.BigEndian.PutUint64(bits[filterBytes+0:], uint64(ngram))
	binary.BigEndian.PutUint64(bits[filterBytes+8:], uint64(skip))
	binary.BigEndian.PutUint64(bits[filterBytes+16:], uint64(numhash))

	for i := 0; i < len(org); i += skip {
		word := org[i:]
		if len(word) > ngram {
			word = word[:ngram]
		}
		if len(word) < ngram {
			break
		}
		hll.Add(uint32(Digest(word)))

		if i/skip%10000 == 0 {
			/* max cardinality를 넘으면,
			 * 다음 개행까지만 처리하고, 나머지는 남김 */
			if hll.Count() > uint64(maxCardinality) {
				nextLineIndex := bytes.IndexByte(org, '\n')
				if nextLineIndex > ngram+skip {
					/* 다음 개행이 ngram, skip 만큼은 떨어져 있어야,
					 * 바로 밑의 bit 처리 코드에 영향이 없음 */
					org = org[:nextLineIndex]
				}
			}
		}

		copy(buf, word)
		for k := 0; k < numhash; k++ {
			buf[ngram] = byte(k)
			digest := Digest(buf) % uint32(filterBytes*8)

			bits[digest>>3] |= 1 << byte(digest&7)
		}
	}
	cardinality := hll.Count()
	binary.BigEndian.PutUint64(bits[filterBytes+24:], uint64(cardinality))
	return int64(len(org)), &NBIndex{bits, ngram, skip, numhash, cardinality}, nil
}

func Load(org []byte) NBIndex {
	cursor := len(org) - tailSize
	ngram := int(binary.BigEndian.Uint64(org[cursor+0:]))
	skip := int(binary.BigEndian.Uint64(org[cursor+8:]))
	numhash := int(binary.BigEndian.Uint64(org[cursor+16:]))
	cardinality := binary.BigEndian.Uint64(org[cursor+24:])
	return NBIndex{org, ngram, skip, numhash, cardinality}
}

package nbstore

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.daumkakao.io/tscoke/nbgrepd/daemon/nbindex"
	"io"
	"log"
	"sync"
	"time"
)

type NBStoreOption struct {
	/* Filtering 조건 */
	/* nbindex등록 안된 크기가 아래보다 커지면, filter에 설정 */
	FilteringThreshold int64
	/* 아래 시간 이상 갱신이 없으면, 그냥 Filtering 해버림 */
	FilteringDuration time.Duration

	/* nbindex 관련 설정 */
	Ngram          int /* 몇글자씩 digest 할지 */
	Skip           int /* ngram을 얼마나 촘촘히 할지 */
	Numhash        int /* bloom filter에서 hash 몇개 만들지*/
	MaxCardinality int /* 최대 cardinality 값 */
	NumBytes       int /* nbindex하나의 크기 */

	Targets []string
}

type NBStore struct {
	bb *BBoltStore

	opt NBStoreOption

	prevFileMap  FileMap
	fileMapMutex sync.Mutex
}

func DefaultOption() NBStoreOption {
	return NBStoreOption{
		FilteringThreshold: 16 * 1024 * 1024,
		FilteringDuration:  time.Minute,
		Ngram:              8,
		Skip:               4,
		Numhash:            7,
		MaxCardinality:     256 * 1024,
		NumBytes:           256 * 1024 * 2,
		Targets:            []string{"./*.log"},
	}

}

/* 모든 곳에서 검색함 */
func New(dbpath string, opt NBStoreOption) (nbs *NBStore, err error) {
	bb, err := NewBBoltStore(dbpath)
	if err != nil {
		return nil, err
	}

	nbs = &NBStore{bb: bb, opt: opt, prevFileMap: make(FileMap)}
	err = nbs.recovery()
	log.Println("[nbstore.Recoery] End ", err)
	if err != nil {
		return nil, err
	}

	return nbs, nil
}

func (nb *NBStore) Close() {
	nb.bb.db.Close()
	nb.bb = nil
}

/* 저장된 bbolt를 바탕으로 NBStore 구조체의 정보를 복구함 */
func (nb *NBStore) recovery() error {
	now := time.Now()

	nb.fileMapMutex.Lock()
	defer nb.fileMapMutex.Unlock()

	/* 저장된 nbindex를 순회하면서, prevFileMap을 재구축한다 */
	return nb.bb.scan(func(k, v []byte) (goNext bool, err error) {
		chunkkey, err := ParseChunkkey(string(k))
		if err != nil { /* 잘못된 키가 있음 */
			return false, errors.Wrapf(err, "Recovery(%s)", k)
		}

		nbf := nbindex.Load(v)

		oldinfo, ok := nb.prevFileMap[chunkkey.fn]
		if !ok {
			oldinfo.Name = chunkkey.fn
			oldinfo.LastTime = now
		}
		if oldinfo.Size < chunkkey.endOff {
			oldinfo.Size = chunkkey.endOff
		}
		log.Println("[nbstore.Recoery] update ", string(k), oldinfo, nbf.String())
		nb.prevFileMap[chunkkey.fn] = oldinfo
		return true, nil
	})
}

/* target들을 순회하여, append된 데이터들에 대해 NBIndex를 추가함 */
func (nb *NBStore) Update() (updateCount, remainBytes int, err error) {
	/* 현재 대상파일들의 상태를 가져옴 */
	fm, err := MakeFileMap(nb.opt.Targets)
	if err != nil {
		return updateCount, remainBytes, errors.Wrap(err, "NBStore.Update")
	}

	now := time.Now()
	for k, info := range fm {
		nb.fileMapMutex.Lock()
		oldinfo, ok := nb.prevFileMap[k]
		if !ok { /* 없으면, 없는데로 정보 만들어줌 */
			oldinfo.Name = k
			oldinfo.Size = 0
			oldinfo.LastTime = now
			nb.prevFileMap[k] = oldinfo
		}
		nb.fileMapMutex.Unlock()

		/* 임계치 이상으로 log가 추가되었거나,
		 * 추가가 된 후, 임계치 이상 시간이 흐른 경우 */
		if nb.opt.FilteringThreshold < info.Size-oldinfo.Size ||
			(oldinfo.Size < info.Size && nb.opt.FilteringDuration < now.Sub(oldinfo.LastTime)) {
			processedSize, chunkkey, err := nb.addLines(k, oldinfo.Size, info.Size)
			if processedSize <= 0 {
				log.Printf("Update OverSize:%v OldFile:%v file:%s => %d, %v",
					nb.opt.FilteringThreshold < info.Size-oldinfo.Size,                                 /* 크기가 커져서 갱신된 경우  */
					(oldinfo.Size < info.Size && nb.opt.FilteringDuration < now.Sub(oldinfo.LastTime)), /*오래되서 갱신한 경우 */
					k,
					processedSize, err)

			} else {
				log.Printf("Update OverSize:%v OldFile:%v file:%s => %d, %s, %v",
					nb.opt.FilteringThreshold < info.Size-oldinfo.Size,                                 /* 크기가 커져서 갱신된 경우  */
					(oldinfo.Size < info.Size && nb.opt.FilteringDuration < now.Sub(oldinfo.LastTime)), /*오래되서 갱신한 경우 */
					k,
					processedSize, chunkkey, err)
			}

			if err != nil {
				return updateCount, remainBytes, errors.Wrapf(err, "NBStore.Update.File(%s)", k)
			}

			nb.fileMapMutex.Lock()
			/* 갱신됨에 따라 파일 크기, 확인한 수정시간 업데이트 */
			org := nb.prevFileMap[k]
			org.Size = oldinfo.Size + int64(processedSize)
			org.LastTime = now

			remainBytes += int(info.Size - oldinfo.Size)

			nb.prevFileMap[k] = org
			nb.fileMapMutex.Unlock()

			updateCount++
		}
	}

	/* 기존의 모든 파일들을 추출함*/
	filenames := []string{}
	nb.fileMapMutex.Lock()
	for k := range nb.prevFileMap {
		filenames = append(filenames, k)
	}
	nb.fileMapMutex.Unlock()

	for _, k := range filenames {
		/* 사라진 파일이면, 삭제함 */
		if _, ok := fm[k]; !ok {
			/* 대상 파일들 탐색 */
			prefix := fmt.Sprintf("%s_%s",
				GetNameTime(k).Format(ChunkkeyTimeFormat),
				k)
			keys := []string{}
			nb.bb.prefixScan([]byte(prefix), func(k, v []byte) (goNext bool, err error) {
				keys = append(keys, string(k))
				return true, nil
			})
			for _, key := range keys {
				err = nb.del(key)
				log.Printf("Delete file %s => %s,%v", k, key, err)
				if err != nil {
					return updateCount, remainBytes, errors.Wrapf(err, "NBStore.Update.DeleteFile(%s)", k)
				}
			}
			nb.fileMapMutex.Lock()
			delete(nb.prevFileMap, k)
			nb.fileMapMutex.Unlock()
			updateCount++
		}
	}

	return updateCount, remainBytes, nil
}

func (nb *NBStore) addLines(fn string, beginOff, endOff int64) (processedSize int64, chunkkey Chunkkey, err error) {
	b := GetBufpool().Get()
	defer bufpool.Put(b)

	ba, err := ReadFilePartially(fn, beginOff, endOff, b)
	if err != nil {
		return 0, chunkkey, errors.Wrap(err, "NBStore.addLines")
	}

	if len(ba) <= 0 {
		return 0, chunkkey, io.EOF
	}

	pos := bytes.LastIndex(ba, []byte("\n"))
	if pos < 0 { /* 개행이 한줄도 없음. 즉 할게 없음 */
		return 0, chunkkey, nil
	}
	/* 마지막 개행 까지만 작업한다.
	 * 예)
	 *   aaaaa\nbbbbb\nccccc
	 *               ^---
	 * cccc부분은 line이 잘렸다고 판단함 */
	cur := ba[:pos]

	processedSize, nbf, err := nbindex.Create(
		cur, nb.opt.Ngram, nb.opt.Skip, nb.opt.Numhash,
		nb.opt.NumBytes, nb.opt.MaxCardinality)
	if err != nil {
		return 0, chunkkey, errors.Wrap(err, "NBStore.addLines")
	}

	if processedSize == 0 {
		return 0, chunkkey, nil
	}

	/* 실제 Filtering한 만큼으로 end offset 조정 */
	endOff = beginOff + processedSize

	key := Chunkkey{GetNameTime(fn), beginOff, endOff, fn}
	err = nb.bb.set([]byte(key.String()), nbf.GetBody())
	return processedSize, key, err

}

func (nb *NBStore) del(fn string) error {
	return nb.bb.del([]byte(fn))
}

func makeItr(chunkkeys *[]Chunkkey, keywords []string) Iterator {
	return func(k, v []byte) (goNext bool, err error) {
		// log.Println("Check ", string(k))
		nbf := nbindex.Load(v)
		for _, keyword := range keywords {
			exp := nbf.CheckString(keyword)
			if !exp { /* false가 하나라도 있으면 못찾은거 */
				return true, nil
			}
		}
		chunkkey, err := ParseChunkkey(string(k))
		if err != nil { /* 잘못된 키가 있음 */
			return false, err
		}
		(*chunkkeys) = append((*chunkkeys), chunkkey)
		return true, nil
	}
}

func (nb *NBStore) Check(begin, end time.Time, keywords []string) (chunkkeys []Chunkkey, err error) {
	min := []byte(begin.Format(ChunkkeyTimeFormat))
	max := []byte(end.Format(ChunkkeyTimeFormat))
	return chunkkeys, nb.bb.rangeScan(min, max, makeItr(&chunkkeys, keywords))
}
func (nb *NBStore) CheckAll(keywords []string) (chunkkeys []Chunkkey, err error) {
	return chunkkeys, nb.bb.scan(makeItr(&chunkkeys, keywords))
}

func (nb *NBStore) GetTailChunk(modifyGapSec, gapSize int64) (chunkkeys []Chunkkey, err error) {
	fm, err := MakeFileMap(nb.opt.Targets)
	if err != nil {
		return nil, err
	}
	/* 이 시간 이후에 수정된적 있는 파일만 tail */
	guideline := time.Now().Add(-time.Duration(modifyGapSec) * time.Second)
	for k, v := range fm {
		if v.LastTime.After(guideline) {
			/* 최근에 업데이트 되었으면 */
			beginOff := v.Size - gapSize
			if beginOff < 0 {
				beginOff = 0
			}
			chunkkeys = append(chunkkeys,
				Chunkkey{beginOff: beginOff,
					endOff: v.Size,
					fn:     k})
		}
	}
	return chunkkeys, nil
}

/* prevFileMap에는 파일의 어느정도까지 필터 되었는지가 들어 있으니,
 * 그 이후의, 필터 안된 영역에 대해 반환한다 */
func (nb *NBStore) GetUnfilteredChunk() (chunkkeys []Chunkkey, err error) {
	fm, err := MakeFileMap(nb.opt.Targets)
	if err != nil {
		return nil, errors.Wrap(err, "NBStore.recovery.Update")
	}

	nb.fileMapMutex.Lock()
	for k, curInfo := range fm {
		prevInfo := nb.prevFileMap[k]
		if prevInfo.Size < curInfo.Size {
			/* filtering 한 영역들 빼고, chunk로 묶음 */
			chunkkeys = append(chunkkeys,
				Chunkkey{beginOff: prevInfo.Size,
					endOff: curInfo.Size,
					fn:     k})
		}
	}
	nb.fileMapMutex.Unlock()
	return chunkkeys, nil
}

func (nb *NBStore) DumpFilemap() FileMap {
	res := make(FileMap)
	nb.fileMapMutex.Lock()
	for k, v := range nb.prevFileMap {
		res[k] = v
	}
	nb.fileMapMutex.Unlock()
	return res
}

func (nb *NBStore) DumpChunk() (res []string, err error) {
	err = nb.bb.scan(func(k, v []byte) (goNext bool, err error) {
		res = append(res, string(k))
		return true, nil
	})
	return res, err
}

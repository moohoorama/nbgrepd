package daemon

import (
	"sync"
	"time"
)

type ChildsInfo struct {
	expire int64

	/* 특정 클러스터에 속한 서버 목
	 * cluster -> [child1,child2,child3...] */
	addrMap map[string][]string
	/* addr -> registtime
	 * 해당 child의 유효시간을 표현하기 위함 */
	registTSMap map[string]int64
	childMutex  sync.Mutex
}

func NewChildsInfo(expire int64) *ChildsInfo {
	return &ChildsInfo{
		expire:      expire,
		addrMap:     make(map[string][]string),
		registTSMap: make(map[string]int64)}
}

func (ci *ChildsInfo) AddChild(cluster string, addr string) {
	ci.childMutex.Lock()
	defer ci.childMutex.Unlock()

	_, ok := ci.registTSMap[addr]
	if !ok { /* dedup */
		ci.addrMap[cluster] = append(ci.addrMap[cluster], addr)
	}
	ci.registTSMap[addr] = time.Now().Unix()
}

func (ci *ChildsInfo) GetChilds(cluster string) (addrs []string) {
	ci.childMutex.Lock()
	defer ci.childMutex.Unlock()

	now := time.Now().Unix()

	for _, addr := range ci.addrMap[cluster] {
		regist := ci.registTSMap[addr]
		if now-regist < ci.expire {
			addrs = append(addrs, addr)
		}
	}
	return addrs
}

func (ci *ChildsInfo) Dump() map[string][]string {
	ci.childMutex.Lock()
	defer ci.childMutex.Unlock()

	now := time.Now().Unix()
	res := make(map[string][]string)
	for cluster := range ci.addrMap {
		addrs := []string{}
		for _, addr := range ci.addrMap[cluster] {
			regist := ci.registTSMap[addr]
			if now-regist < ci.expire {
				addrs = append(addrs, addr)
			}
		}
		res[cluster] = addrs

	}
	return res
}

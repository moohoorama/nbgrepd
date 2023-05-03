package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.daumkakao.io/tscoke/nbgrepd/daemon/env"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type grepForm struct {
	/* Web으로 간단하게 호출할때 씀 */
	Keyword []string  `form:"keyword"      default:""`
	StartTS time.Time `form:"start"        time_format:"2006-01-02"`
	EndTS   time.Time `form:"end"          time_format:"2006-01-02"`
}

type grepResponse struct {
	Response string `json:"response",omitempty"`
	Error    error  `json:"error",omitempty"`

	FilterKeyword        []string            `json:"FilterKeyword",omitempty"`
	GrepForm             grepForm            `json:"GrepForm",omitempty"`
	FilteredChunkCount   int                 `json:"filteredChunkCount,omitempty"`
	UnfilteredChunkCount int                 `json:"unfilteredChunkCount,omitempty"`
	ValidChunkCount      int                 `json:"validChunkCount,omitempty"`
	Data                 map[string][]string `json:"data,omitempty"`
}

/* grep을 수행함
 * Prameter Query로 받는게 기본이고,
 * 여러 문자열을 & 조건으로 검색하고 싶으면, json data로 오는 경우도 받아들임 */
func grepAPI(c *gin.Context) {
	gf := grepForm{}
	gf.EndTS = maxTimeTS

	if err := c.Bind(&gf); err != nil {
		c.JSON(http.StatusBadRequest,
			grepResponse{Response: "bind fail",
				Error: err})
		return
	}

	if len(gf.Keyword) <= 0 {
		c.JSON(http.StatusBadRequest,
			grepResponse{Response: "no keyword"})
		return
	}

	/* FilterIndex에 사용될 Keyword를 추출함 */
	filterKeyword := []string{}

	conf := env.Get()
	minFilterKeywordSize := conf.Ngram + conf.Skip
	for _, keyword := range gf.Keyword {
		if len(keyword) >= minFilterKeywordSize {
			filterKeyword = append(filterKeyword, keyword)
		}
	}
	if len(filterKeyword) <= 0 {
		c.JSON(http.StatusBadRequest,
			grepResponse{Response: fmt.Sprintf("no filter keyword(%s)<%d", gf.Keyword, minFilterKeywordSize)})
		return
	}

	/* bloomfilter로 대상 chunk들을 알아냄 */
	chunkkeys, err := nbStore.Check(gf.StartTS, gf.EndTS, filterKeyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			grepResponse{
				Response: "check fail",
				Error:    err,
				GrepForm: gf,
			})
		return
	}
	filteredChunkCount := len(chunkkeys)

	if filteredChunkCount > conf.MaxFilteredChunkCount {
		c.JSON(http.StatusBadRequest,
			grepResponse{
				FilterKeyword:      filterKeyword,
				GrepForm:           gf,
				FilteredChunkCount: filteredChunkCount,
				Response:           fmt.Sprintf("Too large area(%d<%d)", conf.MaxFilteredChunkCount, filteredChunkCount),
			})
		return
	}

	unfilteredChunkkey, err := nbStore.GetUnfilteredChunk()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			grepResponse{
				Response: "FileScanFail",
				Error:    err,
				GrepForm: gf,
			})
		return
	}

	unfilteredChunkCount := len(unfilteredChunkkey)

	chunkkeys = append(chunkkeys, unfilteredChunkkey...)

	validCount := 0
	resArr := make(map[string][]string)
	for _, chunkkey := range chunkkeys {
		res, err := chunkkey.Grep(gf.Keyword)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				grepResponse{
					Response: "grep fail " + chunkkey.String() + " " + err.Error(),
					Error:    err,
					GrepForm: gf,
				})
			return
		}
		if len(res) > 0 {
			/* filtering된 chunk에 실제로 데이터가 있는 경우 */
			validCount++
			fn := chunkkey.GetFilename()
			resArr[fn] = append(resArr[fn], res...)
		}
	}
	c.JSON(http.StatusOK, grepResponse{
		Response:             "ok",
		FilterKeyword:        filterKeyword,
		GrepForm:             gf,
		FilteredChunkCount:   filteredChunkCount,
		UnfilteredChunkCount: unfilteredChunkCount,
		ValidChunkCount:      validCount,
		Data:                 resArr})
}

func grepallAPI(c *gin.Context) {
	cluster := c.Param("cluster")
	addrs := childsInfo.GetChilds(cluster)

	gf := grepForm{}
	if err := c.Bind(&gf); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(gf.Keyword) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "NullKeyword"})
		return
	}

	conf := env.Get()
	values := url.Values{"keyword": gf.Keyword,
		"start": []string{gf.StartTS.Format("2006-01-02")},
		"end":   []string{gf.EndTS.Format("2006-01-02")}}

	mutex := sync.Mutex{}
	grMap := make(map[string]grepResponse)
	wg := sync.WaitGroup{}
	for _, addr := range addrs {
		addr := addr
		uri := "http://" + addr + "/grep?" + values.Encode()
		wg.Add(1)
		go func() {
			resCode, bodyReader, err := HttpCallWithContext(
				c.Request.Context(),
				http.MethodGet,
				uri,
				conf.HttpTimeoutSec)
			gr := grepResponse{}
			if err == nil {
				var body []byte
				body, err = ioutil.ReadAll(bodyReader)
				if err == nil {
					err = json.Unmarshal(body, &gr)
				}
			}
			if err == nil && resCode/100 != 2 {
				err = fmt.Errorf("statuscode:%d", resCode)
			}
			gr.Error = err

			mutex.Lock()
			grMap[addr] = gr
			mutex.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()

	c.JSON(http.StatusOK, grMap)
}

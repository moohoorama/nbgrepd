package daemon

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.daumkakao.io/tscoke/nbgrepd/daemon/env"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

/* 갱신된지 30분이 안된, 최근 log만 가지고 grep함
 * filter index 타지 않고 수행함 */
func tailGrepAPI(c *gin.Context) {
	gf := grepForm{}
	if err := c.Bind(&gf); err != nil {
		c.JSON(http.StatusBadRequest,
			grepResponse{Response: "bind fail",
				Error: err})
		return
	}

	conf := env.Get()
	chunkkeys, err := nbStore.GetTailChunk(conf.TailModifyGapSec, conf.TailGapSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			grepResponse{
				Response: "GetTailChunkFail ",
				Error:    err,
				GrepForm: gf,
			})
		return
	}

	validCount := 0
	resArr := make(map[string][]string)
	for _, chunkkey := range chunkkeys {
		res, err := chunkkey.Grep(gf.Keyword)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				grepResponse{
					Response: "tail fail " + chunkkey.String() + " " + err.Error(),
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
		GrepForm:             gf,
		UnfilteredChunkCount: len(chunkkeys),
		ValidChunkCount:      validCount,
		Data:                 resArr})
}

func tailGrepallAPI(c *gin.Context) {
	cluster := c.Param("cluster")
	addrs := childsInfo.GetChilds(cluster)

	gf := grepForm{}
	if err := c.Bind(&gf); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conf := env.Get()
	values := url.Values{"keyword": gf.Keyword}

	mutex := sync.Mutex{}
	grMap := make(map[string]grepResponse)
	wg := sync.WaitGroup{}
	for _, addr := range addrs {
		addr := addr
		uri := "http://" + addr + "/tailgrep?" + values.Encode()
		wg.Add(1)
		go func() {
			resCode, bodyReader, err := HttpCallWithContext(
				c.Request.Context(),
				http.MethodGet,
				uri,
				conf.HttpTimeoutSec)
			gr := grepResponse{}
			if resCode/100 == 2 && err == nil {
				var body []byte
				body, err = ioutil.ReadAll(bodyReader)
				if err == nil {
					err = json.Unmarshal(body, &gr)
				}
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

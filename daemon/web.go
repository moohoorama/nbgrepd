package daemon

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func bootupWeb(r *gin.Engine, routePrefix string) {
	r.Static(routePrefix+"/static", "./build/static")

	raw, err := os.ReadFile("./build/index.html")
	if err != nil {
		panic(err)
	}
	reactEntryFile := bytes.Replace(
		raw,
		[]byte(`window.ROUTE_PREFIX=""`),
		[]byte(fmt.Sprintf(`window.ROUTE_PREFIX="%s"`, routePrefix)),
		1,
	)
	reactEntryFile = bytes.ReplaceAll(
		reactEntryFile,
		[]byte("/static/"), []byte(fmt.Sprintf("%s/static/", routePrefix)),
	)

	r.Use(func(c *gin.Context) {
		fmt.Println("bootupWeb Start", c.Writer.Written())
		if c.Writer.Written() {
			return
		}

		c.Data(http.StatusOK, gin.MIMEHTML, reactEntryFile)
	})
}

func clusterInfoGet(c *gin.Context) {
	ci := childsInfo.Dump()

	res := []interface{}{}
	for name, servers := range ci {
		res = append(res,
			map[string]string{
				"name":    name,
				"servers": fmt.Sprintf("%+v", servers),
			})
	}

	c.JSON(http.StatusOK, res)
}

func webGet(c *gin.Context) {

}

package daemon

import (
	"context"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.daumkakao.io/tscoke/nbgrepd/daemon/env"
	"github.daumkakao.io/tscoke/nbgrepd/daemon/nbstore"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	/* 2100년 1월 1일, 즉 먼 미래 */
	maxTimeTS  = time.Unix(4102484400, 0)
	childsInfo *ChildsInfo
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func RunForever() {
	conf := env.Get()
	nbstore.SetBytePool(conf.BufferPoolSize, int(conf.FilteringThreshold))

	childsInfo = NewChildsInfo(conf.ChildExpireSec)
	err := backgroundProcessing()
	if err != nil {
		log.Fatal("backgroundProcessing : ", err)
	}

	srv := run()
	Wait(srv)
}

func Wait(srv *http.Server) {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	Shutdown(srv)
}

func Shutdown(srv *http.Server) {
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("End.")
}

func run() *http.Server {
	conf := env.Get()

	r := gin.New()
	/*
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
	*/

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	if len(conf.BasicAuthName) > 0 && len(conf.BasicAuthPass) > 0 {
		r.Use(gin.BasicAuth(gin.Accounts{
			conf.BasicAuthName: conf.BasicAuthPass,
		}))
	}

	pprof.Register(r)

	group := r.Group(conf.RoutePrefix)

	group.GET("/", rootGet) /* 살아있는지 체크용 */

	group.GET("/web", webGet)                 /* web site */
	group.GET("/clusterinfo", clusterInfoGet) /* cluster, server 목록 */

	group.GET("/grep", grepAPI)                        /* 검색용 api */
	group.GET("/grepall/:cluster", grepallAPI)         /* child들의 grep한 결과를 취합하여 보냄 */
	group.GET("/tailgrep", tailGrepAPI)                /* 검색용 api */
	group.GET("/tailgrepall/:cluster", tailGrepallAPI) /* child들의 grep한 결과를 취합하여 보냄 */
	group.POST("/regist/:cluster/:addr", registGet)    /* 외부의 nbgrepd를 등록함  */
	group.GET("/status", statusGet)                    /* filter 정보 반환 */

	bootupWeb(r, conf.RoutePrefix)

	metricExportHandler := promhttp.Handler()
	group.GET("/metrics", func(c *gin.Context) {
		metricExportHandler.ServeHTTP(c.Writer, c.Request)
	})

	srv := &http.Server{Addr: conf.Bind, Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return srv
}

func rootGet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"response": "ok"})
}

func registGet(c *gin.Context) {
	cluster := c.Param("cluster")
	addr := c.Param("addr")

	childsInfo.AddChild(cluster, addr)

	c.JSON(http.StatusOK, gin.H{"response": "ok", "regist": map[string]string{"addr": addr, "cluster": "cluster"}})
}

func statusGet(c *gin.Context) {
	cks, err := nbStore.DumpChunk()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	filemap := nbStore.DumpFilemap()
	c.JSON(http.StatusOK, gin.H{"Chunk": cks, "Filemap": filemap, "ChildInfo": childsInfo.Dump()})
}

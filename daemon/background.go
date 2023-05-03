package daemon

import (
	"context"
	"github.daumkakao.io/tscoke/nbgrepd/daemon/env"
	"github.daumkakao.io/tscoke/nbgrepd/daemon/nbstore"
	"log"
	"net/http"
	"strings"
	"time"
)

/* 마지막으로 확인한 로그 파일 및 크기들 */

var (
	nbStore *nbstore.NBStore
)

func backgroundProcessing() (err error) {
	conf := env.Get()
	opt := nbstore.NBStoreOption{
		FilteringThreshold: conf.FilteringThreshold,
		FilteringDuration:  time.Duration(conf.FilteringDurationSec) * time.Second,
		Ngram:              conf.Ngram,
		Skip:               conf.Skip,
		Numhash:            conf.Numhash,
		MaxCardinality:     conf.MaxCardinality,
		NumBytes:           conf.NumBytes,
		Targets:            strings.Split(conf.TargetPath, ","),
	}

	log.Println("Bootup - ", conf.DBPath)
	nbStore, err = nbstore.New(conf.DBPath, opt)
	if err != nil {
		return err
	}

	log.Println("[LogIndexing] Start")
	begin := time.Now()

	var curDuration time.Duration
	var expectedDuration time.Duration

	initRemainBytes := 0 /*  맨 처음 보고된 remain bytes */
	progressBytes := 0
	for {
		updateCount, remainBytes, err := nbStore.Update()

		if initRemainBytes == 0 {
			initRemainBytes = remainBytes
		} else {
			progressBytes = initRemainBytes - remainBytes

			curDuration = time.Since(begin)
			if progressBytes > 0 {
				expectedDuration = time.Duration(curDuration.Seconds()*float64(initRemainBytes)/float64(progressBytes)) * time.Second
			}
		}
		log.Printf("[LogIndexing...] updateCount : %5d   RemainBytes : %10d(-%8d)=%10d  (%s/%s)   err : %v",
			updateCount,
			initRemainBytes,
			progressBytes,
			remainBytes,
			curDuration,
			expectedDuration,
			err)
		if err != nil {
			return err
		}
		if updateCount <= 0 {
			break
		}
	}

	log.Println("Start")
	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(conf.BGProcSec))
		for range ticker.C {
			processing()
		}
	}()

	return nil
}

func processing() {
	conf := env.Get()
	uri := conf.MasterAddr + "/regist/" + conf.ClusterName + "/" + conf.DomainNamePort
	res, _, err := HttpCallWithContext(
		context.Background(),
		http.MethodPost,
		uri,
		conf.HttpTimeoutSec)
	log.Println("Regist ", uri, res, err)

	updateCount, remainBytes, err := nbStore.Update()
	log.Println("updateCount : ", updateCount, "RemainBytes : ", remainBytes, " err : ", err)
}

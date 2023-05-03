package daemon

import (
	"context"
	"github.com/pkg/errors"
	"github.daumkakao.io/tscoke/nbgrepd/daemon/env"
	"io"
	"net"
	"net/http"
	"time"
)

var defaultHttpTransport *http.Transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConnsPerHost:   64,
	MaxIdleConns:          512,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func GetClient(sec int) *http.Client {
	return &http.Client{
		Timeout:   time.Duration(time.Duration(sec) * time.Second),
		Transport: defaultHttpTransport}
}

func HttpCallWithContext(c context.Context, method, dst string, timeoutSec int) (resCode int, body io.ReadCloser, err error) {
	request, err := http.NewRequest(method, dst, nil)
	if err != nil {
		return http.StatusBadRequest, nil, errors.Wrap(err, "http.NewRequest-Error")
	}

	conf := env.Get()
	if len(conf.BasicAuthName) > 0 && len(conf.BasicAuthPass) > 0 {
		request.SetBasicAuth(conf.BasicAuthName, conf.BasicAuthPass)
	}

	client := GetClient(timeoutSec)
	if c != nil {
		request = request.WithContext(c)
	}
	resp, err := client.Do(request)
	if err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "Do-Error")
	}

	return resp.StatusCode, resp.Body, nil
}

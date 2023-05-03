package nbstore

import (
	"path/filepath"
	"time"
)

var nametimeFormats = []string{
	/* tsd log format */
	"stderr_2006_01_02_T15.log",
	"stdout_2006_01_02_T15.log",
	"tsd_2006_01_02_T15.log",
	"qd_2006_01_02_T15.log",
	"management_2006_01_02_T15.log",
	"mgr_2006_01_02_T15.log",
	/* metakage log format */
	"2006_01_02.log",
	"error_2006_01_02.log",
	/* metakage admin */
	"admin_2006_01_02.log",
	/* iod httpd log format */
	"coke_reload_20060102.log",
	"trace_20060102.log",
	"wrapper_2006_01_02.log",
	/* kage log format */
	"2006_01_02__15.log",
}

/* LogFile이름으로 날짜/시간 추측*/
func GetNameTime(name string) (t time.Time) {
	basename := filepath.Base(name)
	for _, nf := range nametimeFormats {
		ts, err := time.Parse(nf, basename)
		if err == nil {
			/* 포맷이 다르면 에러날 수 있음.
			 * 에러난 경우는 무시하고, 성공일 경우만 시간 설정해줌 */
			return ts
		}
	}

	return t
}

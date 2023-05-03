package env

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestEnv(t *testing.T) {
	Convey("Env basic test", t, func(c C) {
		conf := Get()
        So(conf.NumBytes, ShouldEqual, 524288)
		Println(conf.NumBytes)

		/* Change REDI_AUTH => "12345" */
		os.Setenv("NUMBYTES", "12345")
		So(Reload(), ShouldBeNil)

		conf = Get()
		So(conf.NumBytes, ShouldEqual, 12345)
	})
}

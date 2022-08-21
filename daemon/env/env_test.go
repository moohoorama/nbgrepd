package env

import (
    "os"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEnv(t *testing.T) {
	Convey("Env basic test", t, func(c C) {
        conf := Get()
        Println(conf.RediAuth)

        /* Change REDI_AUTH => "12345" */
        os.Setenv("REDI_AUTH", "12345")
        So(Reload(), ShouldBeNil)

        conf = Get()
        So(conf.RediAuth, ShouldEqual, "12345")
    })
}


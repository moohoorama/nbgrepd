package nbstore

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestChunkKey(t *testing.T) {
	Convey("Chunkkey Parsing test", t, func(c C) {
		testkey := Chunkkey{
			time.Unix(1262271600, 0),
			1234,
			5678,
			"/a/b/c/d_e_f.txt"}
		So(testkey.String(), ShouldEqual, "2010-01-01T00:00:00_/a/b/c/d_e_f.txt_1234_5678")
		subkey, err := ParseChunkkey(testkey.String())
		So(err, ShouldBeNil)
		So(testkey.String(), ShouldEqual, subkey.String())
		So(testkey.fn, ShouldEqual, subkey.fn)
		So(testkey.beginOff, ShouldEqual, subkey.beginOff)
		So(testkey.endOff, ShouldEqual, subkey.endOff)
	})
}

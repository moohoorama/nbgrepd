package nbindex

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNBIndex(t *testing.T) {
	Convey("NBIndex basic test", t, func(c C) {
		processedSize, res, err := Create([]byte("1234abcd"), 2 /*ngram*/, 1 /*skip*/, 3 /*hash count*/, 32 /*sisze*/, 16 /*maxCardinality*/)
		So(err, ShouldBeNil)
		So(processedSize, ShouldEqual, 8)
		Println(res)
		So(res.CheckString("123"), ShouldBeTrue)
		So(res.CheckString("234"), ShouldBeTrue)
		So(res.CheckString("34abc"), ShouldBeTrue)
		So(res.CheckString("34 abc"), ShouldBeFalse)

		/* byte array로 된 부분을 불러와서 복구하면, 모양이 같아야 한다 */
		dup := Load(res.bits)
		So(res.Detail(), ShouldEqual, dup.Detail())
	})
	Convey("NBIndex skip 2", t, func(c C) {
		processedSize, res, err := Create([]byte("1234abcd"), 4 /*ngram*/, 2 /*skip*/, 3 /*hash count*/, 32 /*sisze*/, 16 /*maxCardinality*/)
		So(err, ShouldBeNil)
		So(processedSize, ShouldEqual, 8)
		Println(res)
		So(res.CheckString("1234ab"), ShouldBeTrue)
		So(res.CheckString("234abc"), ShouldBeTrue)
		So(res.CheckString("34abcd"), ShouldBeTrue)
		So(res.CheckString("34 abc"), ShouldBeFalse)

		/* byte array로 된 부분을 불러와서 복구하면, 모양이 같아야 한다 */
		dup := Load(res.bits)
		So(res.Detail(), ShouldEqual, dup.Detail())
	})

}

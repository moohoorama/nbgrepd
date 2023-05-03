package nbindex

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAlphanumeric(t *testing.T) {
	Convey("To Alphanumeric test", t, func(c C) {
		org := []string{
			`2022/08/20 16:00:00 testenv.go:93: [25-867335980-logic.chunkflusher-chunkflush-16] [encodeNDedup] Key:958048537456605000_1660974605000 org:1(1e7370460d62ecdd)+33->33(f8dcad962a26eafc) Reflush replace err:<nil>`,
			`2022/08/20 16:00:00 testenv.go:93: [25-867335980-logic.chunkflusher-chunkflush-29] [encodeNDedup] Key:970384351137325000_1660978125000 org:1(31c049d7aa093b21)+2->2(e432dc1c21a41411) Reflush replace err:<nil>`,
			`2022/08/20 16:00:00 testenv.go:93: [25-867335980-logic.chunkflusher-chunkflush-26] [encodeNDedup] Key:932481812854234000_1660967834000 org:16(3b0fd8dbbd939271)+83->83(d8acc3c60d48c24d) Reflush replace err:<nil>`,
		}
		answer := []string{
			`2022 08 20 16 00 00 testenv.go 93 25 867335980 logic.chunkflusher chunkflush 16 encodeNDedup Key 958048537456605000_1660974605000 org 1 1e7370460d62ecdd 33 33 f8dcad962a26eafc Reflush replace err nil`,
			`2022 08 20 16 00 00 testenv.go 93 25 867335980 logic.chunkflusher chunkflush 29 encodeNDedup Key 970384351137325000_1660978125000 org 1 31c049d7aa093b21 2 2 e432dc1c21a41411 Reflush replace err nil`,
			`2022 08 20 16 00 00 testenv.go 93 25 867335980 logic.chunkflusher chunkflush 26 encodeNDedup Key 932481812854234000_1660967834000 org 16 3b0fd8dbbd939271 83 83 d8acc3c60d48c24d Reflush replace err nil`,
		}
		for idx, row := range org {
			So(extractAlphaNumeric(row), ShouldEqual, answer[idx])
		}
	})
}

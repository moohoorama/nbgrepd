package nbstore

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestNBSTore(t *testing.T) {
	Convey("NBStore basic test", t, func(c C) {
		os.Remove("./nbstore.db")
		opt := DefaultOption()
		opt.NumBytes = 4096
		opt.MaxCardinality = 64
		opt.FilteringThreshold = 256
		opt.Targets = []string{"./*.go"}
		nbs, err := New("./nbstore.db", opt)
		So(err, ShouldBeNil)

		for {
			updateCount, remainBytes, err := nbs.Update()
			So(err, ShouldBeNil)
			Println("Update :", updateCount, "RemainBytes : ", remainBytes)
			if updateCount <= 0 {
				break
			}
		}

		for _, keywords := range [][]string{
			{"FilteringThreshold"},
			{"nb.opt.Ngram", "nb.opt.Skip,"},
		} {
			chunkkeys, err := nbs.CheckAll(keywords)
			So(err, ShouldBeNil)
			for _, chunkkey := range chunkkeys {
				res, err := chunkkey.Grep(keywords)
				So(err, ShouldBeNil)

				Println("Found ", keywords, ":", chunkkey, "=>\n", res)
				/* 검색 결과들이 keyword들을 포함한 row여야 함 */
				for _, r := range res {
					for _, keyword := range keywords {
						So(r, ShouldContainSubstring, keyword)
					}
				}
			}
		}

		/* 닫고 다시 염 */
		nbs.Close()
		nbs, err = New("./nbstore.db", opt)
		So(err, ShouldBeNil)

		updateCount, remainBytes, err := nbs.Update()
		So(err, ShouldBeNil)
		/* update할개 없어야함 */
		So(updateCount, ShouldBeZeroValue)
		Println("RemainBytes ", remainBytes)
	})
}

func TestTSDLog(t *testing.T) {
	Convey("NBStore basic test", t, func(c C) {
		// os.Remove("./tsdstore.db")
		opt := DefaultOption()
		opt.Targets = []string{"/home/deploy/log/*.log"}
		nbs, err := New("./tsdstore.db", opt)
		So(err, ShouldBeNil)

		for _, keywords := range [][]string{
			{"2022/08/22 10:00:01"},
			{"1,668383547037681000_1660981681"},
		} {
			chunkkeys, err := nbs.CheckAll(keywords)
			So(err, ShouldBeNil)
			for _, chunkkey := range chunkkeys {
				res, err := chunkkey.Grep(keywords)
				So(err, ShouldBeNil)

				Println("Found ", keywords, ":", chunkkey, "=>\n", res)
				if chunkkey.String() == `2022-08-22T10:00:00_0_69142_/home/deploy/testlog/management_2022_08_22_T10.log` {
					Println("Found ", keywords, ":", chunkkey, "=>\n", res)
				}
				/* 검색 결과들이 keyword들을 포함한 row여야 함 */
				for _, r := range res {
					for _, keyword := range keywords {
						So(r, ShouldContainSubstring, keyword)
					}
				}
			}
		}
	})
}

func TestMgmtLog(t *testing.T) {
	Convey("NBStore basic test", t, func(c C) {
		os.Remove("./tempstore.db")
		opt := DefaultOption()
		opt.FilteringThreshold = 1024
		opt.Targets = []string{"/home/deploy/testlog/*.log"}
		nbs, err := New("./tempstore.db", opt)
		So(err, ShouldBeNil)

		for {
			updateCount, remainBytes, err := nbs.Update()
			So(err, ShouldBeNil)
			Println("Update :", updateCount, "RemainBytes : ", remainBytes)
			if updateCount <= 0 {
				break
			}
		}

		for _, keywords := range [][]string{
			{"2022/08/22 10:00:01"},
		} {
			chunkkeys, err := nbs.CheckAll(keywords)
			So(err, ShouldBeNil)
			for _, chunkkey := range chunkkeys {
				res, err := chunkkey.Grep(keywords)
				So(err, ShouldBeNil)

				Println("Found ", keywords, ":", chunkkey, "=>\n", res)
				/* 검색 결과들이 keyword들을 포함한 row여야 함 */
				for _, r := range res {
					for _, keyword := range keywords {
						So(r, ShouldContainSubstring, keyword)
					}
				}
			}
		}
	})
}

package nbfilter

import (
	"testing"
    "io/ioutil"
    "bytes"
    "time"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNBFilter(t *testing.T) {
	Convey("NBFilter basic test", t, func(c C) {
        res := Create([]byte("1234abcd"), 3/*ngram*/, 3/*hash count*/, 32/*sisze*/)
        Println(res)
        So(res.CheckString("123"), ShouldBeTrue)
        So(res.CheckString("234"), ShouldBeTrue)
        So(res.CheckString("34abc"), ShouldBeTrue)
        So(res.CheckString("34 abc"), ShouldBeFalse)
    })

    Convey("TestLogFile", t, func(c C) {
        files, err := ioutil.ReadDir("./log")
        So(err, ShouldBeNil)

        keyword := "2022-08-16T15:00:41.016"

        testBA := [][]byte{}
        nbfArr := []NBFilter{}
        unit := 1024*1024
        for _, file := range files {
            if file.IsDir() {
                continue
            }
            ba, err := ioutil.ReadFile("./log/"+file.Name())
            So(err, ShouldBeNil)

            for i := 0; i < len(ba); i += unit {
                cur := ba[i:]
                if len(cur) > unit {
                    cur = cur[:unit]
                }
                testBA = append(testBA, cur)
                nbf := Create(cur, 6/*ngram*/, 3/*hash count*/, 8*64*1024/*size*/)
                nbfArr = append(nbfArr, nbf)

                Println(file.Name(), i, nbf.String(), nbf.CheckString("300430715702303890"))
            }
        }
        for i, nbf := range nbfArr {
            exp := nbf.CheckString(keyword)
            act := bytes.Contains(testBA[i], []byte(keyword))
            Printf("[%10d~] expect:%6v actual:%6v  match:%6v \n",
                i, exp, act, exp==act)
        }
        /* Perf Check */
        begin := time.Now()
        for _, nbf := range nbfArr {
            nbf.CheckString(keyword)
        }
        Println(time.Since(begin))
    })
}



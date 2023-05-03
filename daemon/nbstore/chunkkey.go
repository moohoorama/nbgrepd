package nbstore

import (
	"fmt"
	"github.com/pkg/errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	ChunkkeyTimeFormat = "2006-01-02T15:04:05"
)

type Chunkkey struct {
	t        time.Time
	beginOff int64
	endOff   int64
	fn       string
}

var tz *time.Location = time.FixedZone("KST", 9*60*60)

func (k Chunkkey) GetFilename() string {
	return k.fn
}

func (k Chunkkey) String() string {
	fullpath, _ := filepath.Abs(k.fn)
	return fmt.Sprintf("%s_%s_%d_%d",
		k.t.In(tz).Format(ChunkkeyTimeFormat),
		fullpath,
		k.beginOff,
		k.endOff)
}

/* chunkkey에 등록된 fn, begin offset, end offset으로
 * 실제 파일을 읽어서 grep 한다 */
func (k Chunkkey) Grep(keywords []string) (res []string, err error) {
	b := make([]byte, k.endOff-k.beginOff)
	ba, err := ReadFilePartially(k.fn, k.beginOff, k.endOff, b)
	if err != nil {
		return nil, errors.Wrapf(err, "Chunkkey.Grep(%s)", k.fn)
	}

	if k.endOff-k.beginOff > int64(len(ba)) {
		return nil, fmt.Errorf("GrepFail to read chunk(target:%d~%d=%d > readdata:%d)",
			k.beginOff, k.endOff, k.endOff-k.beginOff, len(ba))
	}

	raw := strings.Split(string(ba), "\n")
LOOP:
	for _, row := range raw {
		for _, keyword := range keywords {
			if !strings.Contains(row, keyword) {
				continue LOOP
			}
		}
		res = append(res, row)
	}
	return res, nil
}

func ParseChunkkey(arg string) (Chunkkey, error) {
	spl := strings.Split(arg, "_")

	t, err := time.ParseInLocation(ChunkkeyTimeFormat, spl[0], tz)
	if err != nil {
		return Chunkkey{}, err
	}
	beginOff, err := strconv.Atoi(spl[len(spl)-2])
	if err != nil {
		return Chunkkey{}, err
	}
	endOff, err := strconv.Atoi(spl[len(spl)-1])
	if err != nil {
		return Chunkkey{}, err
	}

	fn := strings.Join(spl[1:len(spl)-2], "_")

	return Chunkkey{t, int64(beginOff), int64(endOff), fn}, nil
}

package nbstore

import (
	"github.com/pkg/errors"
	"os"
)

func ReadFilePartially(fn string, beginOff, endOff int64, b []byte) ([]byte, error) {
	file, err := os.Open(fn)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadFilePartially.Open(%s)", fn)
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		return nil, errors.Wrapf(err, "ReadFilePartially.Stat(%s)", fn)
	}

	if endOff > fileinfo.Size() {
		endOff = fileinfo.Size()
	}

	if endOff <= beginOff {
		return nil, nil
	}

	if len(b) > int(endOff-beginOff) {
		b = b[:endOff-beginOff]
	}

	n, err := file.ReadAt(b, beginOff)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadFilePartially.ReadAt(%s)", fn)
	}
	return b[:n], nil
}

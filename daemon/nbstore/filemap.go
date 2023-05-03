package nbstore

import (
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"time"
)

/* 확인한 로그파일에 대한 정보 */
type FileInfo struct {
	Name     string
	Size     int64
	LastTime time.Time
}

type FileMap map[string]FileInfo

func MakeFileMap(paths []string) (fm FileMap, err error) {
	totalFileMap := make(FileMap)
	for _, path := range paths {
		curFileMap, err := makeFileMap(path)
		if err != nil {
			return nil, errors.Wrapf(err, "MakeFileMap(%s)", path)
		}
		for k, v := range curFileMap {
			totalFileMap[k] = v
		}
	}

	return totalFileMap, nil
}
func makeFileMap(path string) (fm FileMap, err error) {
	files, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}

	fm = make(FileMap)
	for _, file := range files {
		fullpath, err := filepath.Abs(file)
		if err != nil {
			return nil, err
		}

		fileinfo, err := os.Stat(fullpath)
		if err != nil { /*그 사이에 파일이 사라진등의 변화가 있으면 실패할 수 있음 */
			log.Println("makeFileMap error:", err, fullpath)
			continue
		}
		if fileinfo.IsDir() {
			continue
		}
		fm[fullpath] = FileInfo{fullpath, fileinfo.Size(), fileinfo.ModTime()}
	}
	return fm, nil
}

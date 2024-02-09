package runner

import (
	"errors"
	fileutil "github.com/projectdiscovery/utils/file"
	"github.com/wjlin0/uncover/sources"
	"io"
	"os"
)

type CSVWriter struct {
	files *os.File
}

func NewCSVWriter(path string) (*CSVWriter, error) {
	c := &CSVWriter{}
	if path == "" {
		return nil, errors.New("path is empty")
	}

	file, err := fileutil.OpenOrCreateFile(path)
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() == 0 {
		_, err = file.Write([]byte("\xEF\xBB\xBF")) // 解决打开乱码问题 -> utf-8编码
		result := sources.Result{}
		b, _ := result.CSVHeader()
		_, _ = file.Write([]byte(b))
		_, _ = file.Write([]byte("\n"))

	} else {
		// 移动光标
		_, _ = file.Seek(stat.Size(), io.SeekStart)
	}

	return c, nil
}

func (c *CSVWriter) Write(b []byte) (int, error) {

	return c.files.Write(b)
}
func (c *CSVWriter) Close() error {

	return c.files.Close()
}

func (c *CSVWriter) WriteString(data string) {
	_, _ = c.Write([]byte(data))
}

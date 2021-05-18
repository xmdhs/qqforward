package push

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var c = http.Client{Timeout: 20 * time.Second}

func httpGet(url string) ([]byte, string, error) {
	reqs, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("HttpGet: %w", err)
	}
	reqs.Header.Set("Accept", "*/*")
	reqs.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36")
	rep, err := c.Do(reqs)
	if rep != nil {
		defer rep.Body.Close()
	}
	if err != nil {
		return nil, "", fmt.Errorf("HttpGet: %w", err)
	}
	if rep.StatusCode != http.StatusOK {
		return nil, "", Httpgeterr{msg: rep.Status, url: url}
	}
	b, err := readlimit(rep.Body, filedatalimit)
	if err != nil {
		return nil, "", fmt.Errorf("HttpGet: %w", err)
	}
	return b, rep.Header.Get("Content-Type"), nil
}

type Httpgeterr struct {
	msg string
	url string
}

func (h Httpgeterr) Error() string {
	return "not 200: " + h.msg + " " + h.url
}

const filedatalimit = 20000000

var Errimg = errors.New("非图片")

func readlimit(r io.Reader, limit int) ([]byte, error) {
	data := make([]byte, 0)
	for {
		b := make([]byte, 4096)
		i, err := r.Read(b)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("readlimit: %w", err)
		}
		if i == 0 {
			break
		}
		data = append(data, b[:i]...)
		if len(data) > limit {
			return nil, Erroverlimit
		}
	}
	return data, nil
}

var Erroverlimit = errors.New("文件大小超过设定的上限")

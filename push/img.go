package push

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"mime/multipart"
	"net/url"
)

func PostFile(filename string, file []byte, caption, chatid string) (bodyBuf *bytes.Buffer, ContentType string, err error) {
	bodyBuf = &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	defer func() {
		e := bodyWriter.Close()
		if err != nil {
			err = fmt.Errorf("postFile: %w", e)
		}
	}()
	fileWriter, err := bodyWriter.CreateFormFile("document", filename)
	if err != nil {
		return nil, "", fmt.Errorf("postFile: %w", err)
	}
	_, err = io.Copy(fileWriter, bytes.NewReader(file))
	if err != nil {
		return nil, "", fmt.Errorf("postFile: %w", err)
	}
	caption = html.UnescapeString(caption)
	caption = url.QueryEscape(caption)
	err = bodyWriter.WriteField("caption", caption)
	if err != nil {
		return nil, "", fmt.Errorf("postFile: %w", err)
	}
	err = bodyWriter.WriteField("chat_id", chatid)
	if err != nil {
		return nil, "", fmt.Errorf("postFile: %w", err)
	}
	return bodyBuf, bodyWriter.FormDataContentType(), nil
}

func Downloadimg(url string, i int) ([]byte, string, error) {
	var errr error
	for a := 0; a < i; a++ {
		b, c, err := httpGet(url)
		if err != nil {
			errr = err
			if errors.Is(err, Errimg) || errors.Is(err, Erroverlimit) {
				return nil, "", fmt.Errorf("Downloadimg url: %v : %w", url, err)
			}
			log.Println(err)
			continue
		}
		return b, c, nil
	}
	return nil, "", fmt.Errorf("Downloadimg: %w", errr)
}

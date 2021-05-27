package push

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

func (p *PushTg) Push(body []byte, ContentType string, a int) {
	for i := 0; i < a; i++ {
		err := p.push(body, ContentType)
		if err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
}

func (p *PushTg) push(body []byte, ContentType string) error {
	req, err := http.NewRequest("POST", "https://api.telegram.org/"+p.tgkey+"/sendDocument", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	req.Header.Set("Content-Type", ContentType)
	reps, err := c.Do(req)
	if reps != nil {
		defer reps.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	t, err := ioutil.ReadAll(reps.Body)
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	var ok isok
	err = json.Unmarshal(t, &ok)
	if !ok.OK || err != nil {
		return fmt.Errorf("push %v: %w", string(t), ErrPush)
	}
	return nil
}

var ErrPush = errors.New("推送失败")

type isok struct {
	OK bool `json:"ok"`
}

type PushTg struct {
	tgkey string
}

func NewPushTg(key string) PushTg {
	return PushTg{tgkey: key}
}

func (p *PushTg) pushtext(message, chatID string) error {
	message = url.QueryEscape(message)
	msg := "chat_id=" + chatID + "&text=" + message
	req, err := http.NewRequest("POST", "https://api.telegram.org/"+p.tgkey+"/sendMessage", strings.NewReader(msg))
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reps, err := c.Do(req)
	if reps != nil {
		defer reps.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	t, err := ioutil.ReadAll(reps.Body)
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	var ok isok
	err = json.Unmarshal(t, &ok)
	if !ok.OK || err != nil {
		return fmt.Errorf("push %v: %w", string(t), ErrPush)
	}
	return nil
}

func (p *PushTg) aPushtext(message, chatID string, a int) {
	var err error
	for i := 0; i < a; i++ {
		err = p.pushtext(message, chatID)
		if err != nil {
			log.Println(err)
			time.Sleep(15 * time.Second)
			continue
		}
		break
	}
}

func (p *PushTg) Pushtext(message, chatID string, a int) {
	l := Split(message, 4000)
	for _, v := range l {
		p.aPushtext(v, chatID, a)
	}
}

func Split(s string, length int) []string {
	r := []string{}
	b := []byte(s)
	for len(b) > length {
		var n int
		for i := 0; i < len(b) && n < length; i++ {
			_, size := utf8.DecodeRune(b[n:])
			n += size
		}
		r = append(r, string(b[:n]))
		b = b[n:]
	}
	if len(b) != 0 {
		r = append(r, string(b))
	}
	return r
}

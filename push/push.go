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
)

func Push(body []byte, ContentType string, _ int) {
	for {
		err := push(body, ContentType)
		if err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
}

func push(body []byte, ContentType string) error {
	req, err := http.NewRequest("POST", "https://api.telegram.org/"+tgkey+"/sendDocument", bytes.NewReader(body))
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
		fmt.Println(string(t))
		return ErrPush
	}
	return nil
}

var ErrPush = errors.New("推送失败")

type isok struct {
	OK bool `json:"ok"`
}

var tgkey string

func SetTgkey(key string) {
	tgkey = key
}

func pushtext(message, chatID string) error {
	message = url.QueryEscape(message)
	msg := "chat_id=" + chatID + "&text=" + message
	req, err := http.NewRequest("POST", "https://api.telegram.org/"+tgkey+"/sendMessage", strings.NewReader(msg))
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
	json.Unmarshal(t, &ok)
	if !ok.OK {
		return Pusherr
	}
	return nil
}

var Pusherr = errors.New("推送失败")

func Pushtext(message, chatID string, _ int) {
	for {
		err := pushtext(message, chatID)
		if err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
}

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"html"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xmdhs/qqforward/push"
)

func main() {
	for {
		func() {
			h := http.Header{}
			h.Add("Authorization", "Bearer "+c.WsToken)
			conn, _, err := websocket.DefaultDialer.Dial(c.WsAddress, h)
			if err != nil {
				log.Println(err)
				return
			}
			defer conn.Close()

			cxt, cancel := context.WithCancel(context.Background())
			defer cancel()

			ch := make(chan []byte, 100)

			callbackhub := NewEvent(cxt, ch)

			m := msgMap{m: make(map[int64]chan *message)}

			callbackhub.OnMsg(m.doMsg)
			if c.PrivateMsg {
				d := doPrivate{ch: make(chan *message, 100)}
				callbackhub.OnMsg(d.doPrivateMsg)
			}

			conn.SetReadDeadline(time.Now().Add(pongWait))
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					time.Sleep(1 * time.Second)
					break
				}
				conn.SetReadDeadline(time.Now().Add(pongWait))
				ch <- msg
			}
		}()
	}
}

type doPrivate struct {
	ch chan *message
	sync.Once
}

func (d *doPrivate) doPrivateMsg(cxt context.Context, msg *message) {
	if msg.MsgType != "private" {
		return
	}

	d.Do(func() {
		go func() {
			for {
				select {
				case <-cxt.Done():
					return
				case m := <-d.ch:
					sendMsg(m, c.PrivateMsgTgId)
				}
			}
		}()
	})

	d.ch <- msg
}

type msgMap struct {
	m map[int64]chan *message
}

func (msg *msgMap) doMsg(cxt context.Context, m *message) {
	id, ok := c.QQgroup[strconv.FormatInt(m.GroupID, 10)]
	if !ok {
		return
	}
	ch, ok := msg.m[m.GroupID]
	if !ok {
		ch = make(chan *message, 100)
		msg.m[m.GroupID] = ch
		go func() {
			for {
				select {
				case <-cxt.Done():
					return
				case m := <-ch:
					sendMsg(m, id)
				}
			}
		}()
	}
	ch <- m
}

func sendMsg(m *message, code string) {
	cc := cqcode(m.Message)

	qq := strconv.FormatInt(m.UserID, 10)
	name := m.Sender.Card
	if name == "" {
		name = m.Sender.Nickname
	}
	header := name + "(" + qq + "): "

	for _, cc := range cc {
		switch cc.Type {
		case "text":
			p.Pushtext(header+cc.Data["text"], code, 8)
		case "image", "record":
			if cc.Data["url"] == "" {
				p.Pushtext(header+m.Message, code, 8)
			}
			go pushFile(cc.Data["url"], header, code)
		case "share":
			p.Pushtext(header+cc.Data["url"], code, 8)
		default:
			b, err := json.Marshal(cc)
			if err != nil {
				log.Println(err)
				return
			}
			p.Pushtext(header+string(b), code, 8)
		}
	}
}

func pushFile(url, header, id string) {
	h := push.Split(header, 100)

	b, ctype, err := push.Downloadimg(url, 8)
	if err != nil {
		p.Pushtext(header+url, id, 5)
	}
	l := strings.Split(ctype, "/")
	if len(l) != 2 {
		p.Pushtext(header+url, id, 5)
		return
	}
	filename := tosha256(b) + "." + l[1]
	buff, c, err := push.PostFile(filename, b, h[0], id)
	if err != nil {
		log.Println(err)
		return
	}
	p.Push(buff.Bytes(), c, 5)
}

func tosha256(data []byte) string {
	s := sha256.New()
	s.Write(data)
	b := s.Sum(nil)
	return hex.EncodeToString(b)
}

const (
	pongWait = 60 * time.Second
)

var cqcodeReg = regexp.MustCompile(`\[CQ:.*?\]`)

func cqcode(code string) []acqcode {
	li := cqcodeReg.FindAllStringIndex(code, -1)

	codelist := make([]acqcode, 0, len(li))

	s := 0

	for _, v := range li {
		var text string
		if s < v[0] {
			text = code[s:v[0]]
		}
		cq := code[v[0]:v[1]]
		s = v[1]
		if text != "" {
			codelist = append(codelist, acqcode{Type: "text", Data: map[string]string{"text": html.UnescapeString(text)}})
		}
		codelist = append(codelist, cqcover(cq))
	}
	var text string
	if s < len(code) {
		text = code[s:]
	}
	if text != "" {
		codelist = append(codelist, acqcode{Type: "text", Data: map[string]string{"text": html.UnescapeString(text)}})
	}

	return codelist
}

func cqcover(code string) acqcode {
	c := code[1 : len(code)-1]
	l := strings.Split(c, ",")
	cq := l[0][3:]
	data := map[string]string{}
	for _, v := range l[1:] {
		i := strings.Index(v, "=")
		data[html.UnescapeString(v[:i])] = html.UnescapeString(v[i+1:])
	}
	return acqcode{Type: cq, Data: data}
}

type acqcode struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}

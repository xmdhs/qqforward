package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

			conn.SetReadDeadline(time.Now().Add(pongWait))
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					time.Sleep(1 * time.Second)
					break
				}
				conn.SetReadDeadline(time.Now().Add(pongWait))
				doMsg(msg)
				if c.PrivateMsg {
					doPrivateMsg(msg)
				}
			}
		}()
	}
}

func doPrivateMsg(msg []byte) {
	cond0 := checkMsg(msg)
	if cond0 {
		return
	}
	var m message
	err := json.Unmarshal(msg, &m)
	if err != nil {
		log.Println(err)
		return
	}
	if m.MsgType != "private" {
		return
	}
	sendMsg(m, c.PrivateMsgTgId)
}

func doMsg(msg []byte) {
	cond0 := checkMsg(msg)
	if cond0 {
		return
	}
	var m message
	err := json.Unmarshal(msg, &m)
	if err != nil {
		log.Println(err)
		return
	}
	if m.GroupID != c.QQgroupID {
		return
	}
	sendMsg(m, c.TgCode)
}

func sendMsg(m message, code string) {
	cc := cqcode(m.Message)

	qq := strconv.FormatInt(m.UserID, 10)
	name := m.Sender.Card
	if name == "" {
		name = m.Sender.Nickname
	}
	header := name + "(" + qq + "): "

	for _, cc := range cc {
		switch cc.Type {
		case "text", "reply":
			push.Pushtext(header+cc.Data["text"], code, 8)
		case "image", "record":
			if cc.Data["url"] == "" {
				push.Pushtext(header+m.Message, code, 8)
			}
			go pushFile(cc.Data["url"], header)
		case "share":
			push.Pushtext(header+cc.Data["url"], code, 8)
		default:
			b, err := json.Marshal(cc)
			if err != nil {
				log.Println(err)
				return
			}
			push.Pushtext(header+string(b), code, 8)
		}
	}
}

func checkMsg(msg []byte) bool {
	var e event
	err := json.Unmarshal(msg, &e)
	if err != nil {
		log.Println(err)
		return true
	}
	if e.Type != "message" {
		return true
	}
	return false
}

func pushFile(url, header string) {
	h := push.Split(header, 100)

	b, ctype, err := push.Downloadimg(url, 8)
	if err != nil {
		push.Pushtext(header+url, c.TgCode, 5)

	}
	l := strings.Split(ctype, "/")
	if len(l) != 2 {
		push.Pushtext(header+url, c.TgCode, 5)
		return
	}
	filename := tosha256(b) + "." + l[1]
	buff, c, err := push.PostFile(filename, b, h[0], c.TgCode)
	if err != nil {
		log.Println(err)
		return
	}
	push.Push(buff.Bytes(), c, 5)
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
		s = v[1] + 1
		if text != "" {
			codelist = append(codelist, acqcode{Type: "text", Data: map[string]string{"text": text}})
		}
		codelist = append(codelist, cqcover(cq))
	}
	var text string
	if s < len(code) {
		text = code[s:]
	}
	if text != "" {
		codelist = append(codelist, acqcode{Type: "text", Data: map[string]string{"text": text}})
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
		data[v[:i]] = v[i+1:]
	}
	return acqcode{Type: cq, Data: data}
}

type acqcode struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}

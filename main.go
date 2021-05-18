package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
		h := http.Header{}
		h.Add("Authorization", "Bearer "+c.WsToken)
		conn, _, err := websocket.DefaultDialer.Dial(c.WsAddress, h)
		if err != nil {
			log.Println(err)
			continue
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
		}
	}
}

func doMsg(msg []byte) {
	var e event
	err := json.Unmarshal(msg, &e)
	if err != nil {
		log.Println(err)
		return
	}
	if e.Type != "message" {
		return
	}
	var m message
	err = json.Unmarshal(msg, &m)
	if err != nil {
		log.Println(err)
		return
	}
	if m.GroupID != c.QQgroupID {
		return
	}

	cc := cqcode(m.Message)

	qq := strconv.FormatInt(m.UserID, 10)
	header := m.Sender.Card + "(" + qq + "): "

	for _, cc := range cc {
		switch cc.atype {
		case "text", "reply":
			push.Pushtext(header+cc.data["text"], c.TgCode, 5)
		case "image", "record":
			if cc.data["url"] == "" {
				push.Pushtext(header+m.Message, c.TgCode, 5)
			}
			go pushFile(cc.data["url"], header)

		case "share":
			push.Pushtext(header+cc.data["url"], c.TgCode, 5)

		default:
			push.Pushtext(header + fmt.Sprint(cc), c.TgCode, 5)
		}
	}
}

func pushFile(url, header string) {
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
	buff, c, err := push.PostFile(filename, b, header, c.TgCode)
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
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1000000
)

var cqcodeReg = regexp.MustCompile(`\[CQ:.*?\]`)

func cqcode(code string) []acqcode {
	li := cqcodeReg.FindAllStringIndex(code, -1)

	codelist := make([]acqcode, 0, len(li))

	s := 0

	for _, v := range li {
		text := code[s:v[0]]
		cq := code[v[0]:v[1]]
		s = v[1] + 1
		if text != "" {
			codelist = append(codelist, acqcode{atype: "text", data: map[string]string{"text": text}})
		}
		codelist = append(codelist, cqcover(cq))
	}
	text := code[s:]
	if text != "" {
		codelist = append(codelist, acqcode{atype: "text", data: map[string]string{"text": text}})
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
	return acqcode{atype: cq, data: data}
}

type acqcode struct {
	atype string
	data  map[string]string
}

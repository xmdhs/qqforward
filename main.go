package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xmdhs/qqforward/push"
)

func main() {
	h := http.Header{}
	h.Add("Authorization", "Bearer "+c.WsToken)
	conn, _, err := websocket.DefaultDialer.Dial(c.WsAddress, h)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	cxt := context.Background()
	cxt, cancel := context.WithCancel(cxt)

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	go ping(cxt, conn, cancel)

	defer cancel()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			continue
		}
		go doMsg(msg)
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
	header := m.Sender.Nickname + "(" + qq + ") : "

	switch cc.atype {
	case "text", "reply":
		push.Pushtext(header+cc.data["text"], c.TgCode, 5)
	case "image", "record":
		if cc.data["url"] == "" {
			push.Pushtext(header+m.Message, c.TgCode, 5)
		}
		pushFile(cc.data["url"], header)

	case "share":
		push.Pushtext(header+cc.data["url"], c.TgCode, 5)
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
	filename := tosha256(b)
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

func ping(cxt context.Context, conn *websocket.Conn, cancel context.CancelFunc) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
		cancel()
	}()
	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)
				return
			}
		case <-cxt.Done():
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
	}
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

func cqcode(code string) acqcode {
	if strings.HasPrefix(code, "[") && strings.HasSuffix(code, "]") {
		c := code[1 : len(code)-1]
		l := strings.Split(c, ",")
		cq := l[0][3:]
		data := map[string]string{}
		for _, v := range l[1:] {
			i := strings.Index(v, "=")
			data[v[:i]] = v[i+1:]
		}
		return acqcode{atype: cq, data: data}

	} else {
		return acqcode{atype: "text", data: map[string]string{"text": code}}
	}
}

type acqcode struct {
	atype string
	data  map[string]string
}

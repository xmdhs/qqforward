package main

import (
	"context"
	"strconv"
	"strings"
	"sync"
)

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

func check(msg []string, chatID string) func(cxt context.Context, msg *message) {
	return func(cxt context.Context, m *message) {
		if m.MsgType != "group" {
			return
		}

		for _, v := range msg {
			if strings.Contains(m.Message, v) {
				name := m.Sender.Card
				if name == "" {
					name = m.Sender.Nickname
				}
				p.Pushtext(strconv.FormatInt(m.GroupID, 10)+" : "+name, chatID, 8)
				sendMsg(m, chatID)
			}
		}
	}
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

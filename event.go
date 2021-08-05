package main

import (
	"context"
	"encoding/json"
	"log"
)

type Event struct {
	ch    chan []byte
	funcs map[string]func(cxt context.Context, msg []byte)
	OnMsg func(cxt context.Context, msg *message)
}

func NewEvent(cxt context.Context, ch chan []byte) *Event {
	e := &Event{}
	e.ch = ch
	e.funcs = make(map[string]func(context.Context, []byte))
	e.funcs["message"] = e.onMsg
	go e.do(cxt)
	return e
}

func (e *Event) onMsg(cxt context.Context, msg []byte) {
	m := message{}
	err := json.Unmarshal(msg, &m)
	if err != nil {
		log.Println(err)
		return
	}
	e.OnMsg(cxt, &m)
}

func (e *Event) do(cxt context.Context) {
	for {
		select {
		case <-cxt.Done():
			return
		case b := <-e.ch:
			t := checkMsg(b)
			f, ok := e.funcs[t]
			if !ok {
				continue
			}
			f(cxt, b)
		}
	}
}

func checkMsg(msg []byte) string {
	var e event
	err := json.Unmarshal(msg, &e)
	if err != nil {
		return ""
	}
	return e.Type
}

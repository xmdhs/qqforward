package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
)

type Event struct {
	m       sync.RWMutex
	ch      chan []byte
	msgfunc []func(cxt context.Context, msg *message)

	funcs map[string]func(cxt context.Context, msg []byte)
}

func NewEvent(cxt context.Context, ch chan []byte) *Event {
	e := &Event{}
	e.msgfunc = make([]func(cxt context.Context, msg *message), 0)
	e.ch = ch

	e.funcs = make(map[string]func(cxt context.Context, msg []byte))
	e.funcs["message"] = e.onMsg

	go e.do(cxt)
	return e
}

func (e *Event) OnMsg(callback func(cxt context.Context, msg *message)) {
	e.m.Lock()
	defer e.m.Unlock()
	e.msgfunc = append(e.msgfunc, callback)
}

func (e *Event) onMsg(cxt context.Context, msg []byte) {
	m := message{}
	err := json.Unmarshal(msg, &m)
	if err != nil {
		log.Println(err)
		return
	}
	e.m.RLock()
	defer e.m.RUnlock()
	for _, v := range e.msgfunc {
		v(cxt, &m)
	}
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

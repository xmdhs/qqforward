package main

import (
	"encoding/json"
	"os"

	"github.com/xmdhs/qqforward/push"
)

type config struct {
	TgToken        string
	QQgroup        map[string]string
	WsAddress      string
	WsToken        string
	PrivateMsg     bool
	PrivateMsgTgId string
}

var c config

func readConfig() {
	b, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &c)
	if err != nil {
		panic(err)
	}
}

var p push.PushTg

func init() {
	readConfig()
	p = push.NewPushTg(c.TgToken)
}

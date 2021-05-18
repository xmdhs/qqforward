package main

import (
	"encoding/json"
	"os"

	"github.com/xmdhs/qqforward/push"
)

type config struct {
	TgToken   string
	TgCode    string
	QQgroupID int64
	WsAddress string
	WsToken   string
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

func init() {
	readConfig()
	push.SetTgkey(c.TgToken)
}

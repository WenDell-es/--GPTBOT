package main

import (
	"github.com/go-ini/ini"
	"gptbot/src/botService"
	"gptbot/src/config"
)

func main() {
	iniCfg, err := ini.Load("./configuration/conf.ini")
	if err != nil {
		panic(err)
	}
	cfg := config.Config{}

	if err = iniCfg.MapTo(&cfg); err != nil {
		panic(err)
	}
	botServer := botService.NewBotServer(cfg)
	botServer.Start()
}

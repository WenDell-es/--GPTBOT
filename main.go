// Package main ZeroBot-Plugin main file
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"gptbot/admins"
	"gptbot/log"
	_ "gptbot/store"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	// 以下是相关插件，不要改变插件import的顺序
	_ "gptbot/plugin/madokapicture"

	_ "github.com/FloatTech/ZeroBot-Plugin/console"
	"github.com/FloatTech/ZeroBot-Plugin/kanban"
	//_ "github.com/FloatTech/ZeroBot-Plugin/plugin/b14"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/base64gua"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/choose"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chouxianghua"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/coser"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cpstory"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dailynews"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/diana"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dish"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/drawlots"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/emojimix"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/font"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/fortune"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/funny"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/genshin"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/gif"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/github"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/inject"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/kfccrazythursday"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolicon"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/mcfish"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/omikuji"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qqwife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/realcugan"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/reborn"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/robbery"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/runcode"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/saucenao"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/setutime"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shadiao"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shindan"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tarot"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tiangou"
	//
	_ "gptbot/plugin/spouse"

	// 最低优先级
	_ "gptbot/plugin/gptbot" // chatgpt 机器人
	// 以上是相关插件，不要改变插件import的顺序

	"github.com/FloatTech/ZeroBot-Plugin/kanban/banner"
	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type zbpcfg struct {
	Z zero.Config        `json:"zero"`
	W []*driver.WSClient `json:"ws"`
	S []*driver.WSServer `json:"wss"`
}

var config zbpcfg

func init() {
	sus := make([]int64, 0, 16)
	// 解析命令行参数
	d := flag.Bool("d", false, "Enable debug level log and higher.")
	w := flag.Bool("w", false, "Enable warning level log and higher.")
	h := flag.Bool("h", false, "Display this help.")
	// g := flag.String("g", "127.0.0.1:3000", "Set webui url.")
	// 直接写死 AccessToken 时，请更改下面第二个参数
	token := flag.String("t", "", "Set AccessToken of WSClient.")
	// 直接写死 URL 时，请更改下面第二个参数
	url := flag.String("u", "ws://127.0.0.1:6700", "Set Url of WSClient.")
	// 默认昵称
	adana := flag.String("n", "猫娘", "Set default nickname.")
	prefix := flag.String("p", "/", "Set command prefix.")
	runcfg := flag.String("c", "", "Run from config file.")
	save := flag.String("s", "", "Save default config to file and exit.")
	late := flag.Uint("l", 233, "Response latency (ms).")
	rsz := flag.Uint("r", 4096, "Receiving buffer ring size.")
	maxpt := flag.Uint("x", 4, "Max process time (min).")

	flag.Parse()

	if *h {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}
	if *d && !*w {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if *w {
		logrus.SetLevel(logrus.WarnLevel)
	}

	for _, s := range flag.Args() {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			continue
		}
		sus = append(sus, i)
	}

	// 通过代码写死的方式添加主人账号
	sus = append(sus, admins.Admins...)
	// sus = append(sus, 87654321)

	// 启用 webui
	// go webctrl.RunGui(*g)

	if *runcfg != "" {
		f, err := os.Open(*runcfg)
		if err != nil {
			panic(err)
		}
		config.W = make([]*driver.WSClient, 0, 2)
		err = json.NewDecoder(f).Decode(&config)
		f.Close()
		if err != nil {
			panic(err)
		}
		config.Z.Driver = make([]zero.Driver, len(config.W)+len(config.S))
		for i, w := range config.W {
			config.Z.Driver[i] = w
		}
		for i, s := range config.S {
			config.Z.Driver[i+len(config.W)] = s
		}
		logrus.Infoln("[main] 从", *runcfg, "读取配置文件")
		return
	}
	config.W = []*driver.WSClient{driver.NewWebSocketClient(*url, *token)}
	config.Z = zero.Config{
		NickName:       append([]string{*adana}, "猫娘", "猫猫"),
		CommandPrefix:  *prefix,
		SuperUsers:     sus,
		RingLen:        *rsz,
		Latency:        time.Duration(*late) * time.Millisecond,
		MaxProcessTime: time.Duration(*maxpt) * time.Minute,
		Driver:         []zero.Driver{config.W[0]},
	}

	if *save != "" {
		f, err := os.Create(*save)
		if err != nil {
			panic(err)
		}
		err = json.NewEncoder(f).Encode(&config)
		f.Close()
		if err != nil {
			panic(err)
		}
		logrus.Infoln("[main] 配置文件已保存到", *save)
		os.Exit(0)
	}
	log.LogInit()
}

func main() {
	if !strings.Contains(runtime.Version(), "go1.2") { // go1.20之前版本需要全局 seed，其他插件无需再 seed
		rand.Seed(time.Now().UnixNano()) //nolint: staticcheck
	}
	// 帮助
	zero.OnFullMatchGroup([]string{"/help", "help", "菜单", "-help"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(banner.Banner, "\n发送\"/服务列表\"查看 bot 功能\n发送\"/用法name\"查看功能用法"))
		})
	zero.OnFullMatch("查看zbp公告", zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(strings.ReplaceAll(kanban.Kanban(), "\t", "")))
		})
	zero.RunAndBlock(&config.Z, process.GlobalInitMutex.Unlock)
}

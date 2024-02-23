package madokapicture

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gptbot/store"
	"strconv"
	"time"
)

const (
	Storage     = "storage/"
	Daily       = "daily/"
	SectionName = "bot"
	DefaultPath = "./config/config.ini"
)

var SelfId string

type CosCfg struct {
	Host      string
	SecretID  string
	SecretKey string
}

func init() {
	conf, err := ini.Load(DefaultPath)
	if err != nil {
		logrus.Fatalln(err)
	}
	SelfId = conf.Section(SectionName).Key("Id").String()
	engine := control.Register("yuantu", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "- 来份圆图  随机发一张魔圆的图\n" +
			"- 圆图十连 随机发⑩张魔圆的图\n" +
			"- 查询圆图数量 查询图库中图片数量\n",
		Brief:             "随机发一些圆图",
		PrivateDataFolder: "yuantu",
	}).ApplySingle(ctxext.DefaultSingle)

	pool := NewPicBuffer(
		engine.DataFolder()+"frequency.json",
		time.Hour*6,
		50,
		100,
		"http://127.0.0.1:5700/send_msg",
	)
	pool.Start()

	engine.OnFullMatch("查询圆图数量", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			logrus.WithFields(logrus.Fields{
				"command": "查询圆图数量",
				"userId":  ctx.Event.UserID,
				"groupId": ctx.Event.GroupID,
			}).Infoln()
			objs, err := store.GetStoreClient().FetchAllFileInfo(Storage)
			if err != nil {
				logrus.Errorln("获取对象信息错误", err)
				ctx.SendChain(message.Text("获取对象信息错误 " + err.Error()))
				return
			}
			ctx.SendChain(message.Text("当前图库总数为 " + strconv.Itoa(len(objs))))
		})

	engine.OnFullMatch("来份圆图", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			logrus.WithFields(logrus.Fields{
				"command": "来份圆图",
				"userId":  ctx.Event.UserID,
				"groupId": ctx.Event.GroupID,
			}).Infoln()
			urls := pool.GetUrls(1)
			ctx.SendChain(message.Image(urls[0]))
		})

	engine.OnFullMatch("圆图十连", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			logrus.WithFields(logrus.Fields{
				"command": "圆图十连",
				"userId":  ctx.Event.UserID,
				"groupId": ctx.Event.GroupID,
			}).Infoln()
			if pool.GetBufferCount() < 10 {
				ctx.SendChain(message.Text("正在尽全力补充圆图，请等等喵~"))
			}
			urls := pool.GetUrls(10)
			for _, url := range urls {
				ctx.SendChain(message.Image(url))
			}
		})
}

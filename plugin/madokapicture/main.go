package madokapicture

import (
	"crypto/rand"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gptbot/store"
	"math/big"
	"strconv"
	"sync"
	"time"
)

const (
	Storage = "storage/"
	Daily   = "daily/"
)

type CosCfg struct {
	Host      string
	SecretID  string
	SecretKey string
}

func init() {
	engine := control.Register("yuantu", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "- 来份圆图  随机发一张魔圆的图\n" +
			"- 圆图十连 随机发⑩张魔圆的图\n" +
			"- 今日圆图 发送今天图库里新增的图\n" +
			"- 查询圆图数量 查询图库中图片数量\n",
		Brief:             "随机发一些圆图",
		PrivateDataFolder: "yuantu",
	}).ApplySingle(ctxext.DefaultSingle)
	BufferInit()
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
			urls := buffer.GetUrls(1)
			ctx.SendChain(message.Image(urls[0]))
		})

	engine.OnFullMatch("圆图十连", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			logrus.WithFields(logrus.Fields{
				"command": "圆图十连",
				"userId":  ctx.Event.UserID,
				"groupId": ctx.Event.GroupID,
			}).Infoln()
			urls := buffer.GetUrls(10)
			wg := sync.WaitGroup{}
			for _, url := range urls {
				go func(u string) {
					wg.Add(1)
					ctx.SendChain(message.Image(u))
					wg.Done()
				}(url)
			}
			wg.Wait()
		})

	engine.OnFullMatch("今日圆图", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			logrus.WithFields(logrus.Fields{
				"command": "今日圆图",
				"userId":  ctx.Event.UserID,
				"groupId": ctx.Event.GroupID,
			}).Infoln()
			dir := Daily + time.Now().Format("20060102") + "/"
			objs, err := store.GetStoreClient().FetchAllFileInfo(dir)
			if err != nil {
				logrus.Errorln("获取对象信息错误", err)
				ctx.SendChain(message.Text("获取对象信息错误 " + err.Error()))
				return
			}
			if len(objs) == 0 {
				ctx.SendChain(message.Text("今天图库里还没有新增的图呢~"))
				return
			}
			ctx.SendChain(message.Text("今天图库里一共新增了" + strconv.Itoa(len(objs)) + "张图片呢喵~"))
			for i := 0; i < len(objs); i++ {
				ctx.SendChain(message.Image(store.GetStoreClient().GetObjectUrl(objs[i].Key)))
			}
		})

}

func getRandomNum(n int) int {
	b := new(big.Int).SetInt64(int64(n))
	i, err := rand.Int(rand.Reader, b)
	if err != nil {
		logrus.Errorln(err)
		return -1
	}
	return int(i.Int64())
}

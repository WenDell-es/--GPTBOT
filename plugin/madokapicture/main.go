package madokapicture

import (
	"github.com/FloatTech/floatbox/binary"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gptbot/store"
	"hash/crc64"
	"math/rand"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

type CosCfg struct {
	Host      string
	SecretID  string
	SecretKey string
}

const (
	Storage = "storage/"
	Daily   = "daily/"
)

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

	engine.OnFullMatch("查询圆图数量", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
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
			objs, err := store.GetStoreClient().FetchAllFileInfo(Storage)
			if err != nil {
				logrus.Errorln("获取对象信息错误", err)
				ctx.SendChain(message.Text("获取对象信息错误 " + err.Error()))
				return
			}
			obj := objs[getRandomNum(len(objs), ctx.Event.UserID)]
			ourl := store.GetStoreClient().GetObjectUrl(obj.Key)

			if id := ctx.SendChain(message.Image(ourl)); id.ID() == 0 {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text("【图片发送失败, 请联系维护者】"))
			}

		})

	engine.OnFullMatch("圆图十连", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			objs, err := store.GetStoreClient().FetchAllFileInfo(Storage)
			if err != nil {
				logrus.Errorln("获取对象信息错误", err)
				ctx.SendChain(message.Text("获取对象信息错误 " + err.Error()))
				return
			}
			sum := crc64.New(crc64.MakeTable(crc64.ISO))
			sum.Write(binary.StringToBytes(time.Now().Format("2006-01-02 15:04:05.000")))
			sum.Write((*[8]byte)(unsafe.Pointer(&ctx.Event.UserID))[:])
			r := rand.New(rand.NewSource(int64(sum.Sum64())))
			wg := sync.WaitGroup{}
			for i := 0; i < 10; i++ {
				obj := objs[r.Intn(len(objs))]
				ourl := store.GetStoreClient().GetObjectUrl(obj.Key)
				go func() {
					wg.Add(1)
					ctx.SendChain(message.Image(ourl))
					wg.Done()
				}()
				wg.Wait()
			}

		})

	engine.OnFullMatch("今日圆图", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
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

func getRandomNum(n int, uid int64) int {
	sum := crc64.New(crc64.MakeTable(crc64.ISO))
	sum.Write(binary.StringToBytes(time.Now().Format("2006-01-02 15:04:05.000")))
	sum.Write((*[8]byte)(unsafe.Pointer(&uid))[:])
	r := rand.New(rand.NewSource(int64(sum.Sum64())))
	return r.Intn(n)
}

package madokapicture

import (
	"github.com/FloatTech/floatbox/binary"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"github.com/tencentyun/cos-go-sdk-v5"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"golang.org/x/net/context"
	"hash/crc64"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
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
			//"- 今日圆图 发送今天图库里新增的图\n" +
			"- 查询圆图数量 查询图库中图片数量\n",
		Brief:             "随机发一些圆图",
		PrivateDataFolder: "yuantu",
	}).ApplySingle(ctxext.DefaultSingle)
	cosClient, err := cosClientInit(engine.DataFolder())
	if err != nil {
		logrus.Errorln("cos客户端配置失败", err)
		return
	}

	engine.OnFullMatch("查询圆图数量", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			objs, err := fetchAllFileInfo(cosClient, Storage)
			if err != nil {
				logrus.Errorln("获取对象信息错误", err)
				ctx.SendChain(message.Text("获取对象信息错误 " + err.Error()))
				return
			}
			ctx.SendChain(message.Text("当前图库总数为 " + strconv.Itoa(len(objs))))
		})

	engine.OnFullMatch("来份圆图", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			objs, err := fetchAllFileInfo(cosClient, Storage)
			if err != nil {
				logrus.Errorln("获取对象信息错误", err)
				ctx.SendChain(message.Text("获取对象信息错误 " + err.Error()))
				return
			}
			obj := objs[getRandomNum(len(objs), ctx.Event.UserID)]
			ourl := cosClient.Object.GetObjectURL(obj.Key)

			if id := ctx.SendChain(message.Image(ourl.String())); id.ID() == 0 {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text("【图片发送失败, 请联系维护者】"))
			}

		})

	engine.OnFullMatch("圆图十连", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			objs, err := fetchAllFileInfo(cosClient, Storage)
			if err != nil {
				logrus.Errorln("获取对象信息错误", err)
				ctx.SendChain(message.Text("获取对象信息错误 " + err.Error()))
				return
			}
			sum := crc64.New(crc64.MakeTable(crc64.ISO))
			sum.Write(binary.StringToBytes(time.Now().Format("2006-01-02 15:04:05.000")))
			sum.Write((*[8]byte)(unsafe.Pointer(&ctx.Event.UserID))[:])
			r := rand.New(rand.NewSource(int64(sum.Sum64())))
			for i := 0; i < 10; i++ {
				obj := objs[r.Intn(len(objs))]
				ourl := cosClient.Object.GetObjectURL(obj.Key)
				go func() {
					if id := ctx.SendChain(message.Image(ourl.String())); id.ID() == 0 {
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("【图片发送失败, 请联系维护者】"))
					}
				}()
			}

		})

}

func cosClientInit(dataPath string) (*cos.Client, error) {
	conf, err := ini.Load(dataPath + "conf.ini")
	if err != nil {
		return nil, err
	}
	cosCfg := &CosCfg{}
	err = conf.MapTo(cosCfg)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(cosCfg.Host)
	return cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cosCfg.SecretID,
			SecretKey: cosCfg.SecretKey,
		},
	}), nil
}

func fetchAllFileInfo(cosClient *cos.Client, prefix string) ([]cos.Object, error) {
	var marker string
	opt := &cos.BucketGetOptions{
		Prefix:    prefix,
		Delimiter: "/",
		MaxKeys:   1000,
	}
	res := []cos.Object{}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := cosClient.Bucket.Get(context.Background(), opt)
		if err != nil {
			return nil, err
		}
		res = append(res, v.Contents...)
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}
	res = res[1:]
	return res, nil
}

func getRandomNum(n int, uid int64) int {
	sum := crc64.New(crc64.MakeTable(crc64.ISO))
	sum.Write(binary.StringToBytes(time.Now().Format("2006-01-02 15:04:05.000")))
	sum.Write((*[8]byte)(unsafe.Pointer(&uid))[:])
	r := rand.New(rand.NewSource(int64(sum.Sum64())))
	return r.Intn(n)
}

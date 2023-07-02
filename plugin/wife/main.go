// Package wife 抽老婆
package wife

import (
	"bytes"
	"encoding/json"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"sync"
	"time"
)

type record struct {
	Date time.Time
	Wife wife
}

type wife struct {
	Name         string
	Source       string
	UploaderName string
	UploaderId   int64
}

func init() {
	engine := control.Register("wife", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "- 抽老婆",
		Brief:            "从老婆库抽每日老婆",
		PublicDataFolder: "Wife",
	}).ApplySingle(ctxext.DefaultSingle)
	_ = os.MkdirAll(engine.DataFolder()+"wives", 0755)

	userRecords := sync.Map{}

	engine.OnFullMatch("抽老婆").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var card wife
			if rd, ok := userRecords.Load(ctx.Event.UserID); ok &&
				rd.(record).Date.Format("20060102") == time.Now().Format("20060102") {
				card = rd.(record).Wife
			} else {
				cards, err := getWifeCards(engine.DataFolder() + "wife.json")
				if err != nil {
					logrus.Errorln(err)
					ctx.SendChain(message.At(ctx.Event.UserID), message.Text("喵喵喵！老婆池加载失败了喵~~！", err))
					return
				}
				card = cards[fcext.RandSenderPerDayN(ctx.Event.UserID, len(cards))]
				userRecords.Store(ctx.Event.UserID, record{
					Date: time.Now(),
					Wife: card,
				})
			}

			data, err := os.ReadFile(engine.DataFolder() + "wives/" + card.Name)
			wifeName, _, _ := strings.Cut(card.Name, ".")
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text("今天的二次元老婆是~【", wifeName, "】哒\n【图片下载失败: ", err, "】"))
				return
			}
			if id := ctx.SendChain(message.At(ctx.Event.UserID), message.Text("今天的二次元老婆是~【", wifeName, "】哒!\n来自作品【", card.Source, "】哦~\n上传人是【", card.UploaderName, ",", card.UploaderId, "】呢"), message.ImageBytes(data)); id.ID() == 0 {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text("今天的二次元老婆是~【", wifeName, "】哒\n【图片发送失败, 请联系维护者】"))
			}
		})

	engine.OnFullMatch("添加老婆", zero.MustProvidePicture).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			addWifeEvent, cancel := ctx.FutureEvent("message", ctx.CheckSession()).Repeat()
			defer cancel()

			newWife := wife{}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入新老婆的名称喵~~"))
			name := <-addWifeEvent
			rawName := strings.TrimSpace(name.Event.RawMessage)

			newWife.Name = rawName + ".jpg"
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("接下来请为"+strings.TrimSpace(name.Event.RawMessage)+"添加角色出处哦~"))
			source := <-addWifeEvent
			newWife.Source = source.Event.RawMessage
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("正在录入新老婆信息嗒！\n老婆名字:"+strings.TrimSpace(name.Event.RawMessage)+"\n老婆出处:"+source.Event.RawMessage))
			cards, err := getWifeCards(engine.DataFolder() + "wife.json")
			if len(cards) >= 1000 {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text("老婆数量已经达到最大值1000了，不能再添加了"))
				return
			}
			newWife.UploaderName = ctx.Event.Sender.NickName
			newWife.UploaderId = ctx.Event.UserID
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("喵喵喵！老婆池加载失败了喵~~！", err),
				)
				return
			}
			for _, card := range cards {
				if card.Name == newWife.Name {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("老婆已经存在啦!"))
					return
				}
			}
			cards = append(cards, newWife)

			url := ctx.State["image_url"].([]string)[0]
			picPath := engine.DataFolder() + "wives/" + name.Event.RawMessage + ".jpg"
			err = file.DownloadTo(url, picPath)
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("错误：", err.Error()))
				os.Remove(picPath)
				return
			}
			err = convertPictureToJpg(picPath)
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("错误：", err.Error()))
				os.Remove(picPath)
				return
			}
			err = saveWifeFile(engine.DataFolder()+"wife.json", cards)
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("保存老婆竟然出错了喵！！", err),
				)
				return
			}
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功！"))
		})

	engine.OnFullMatch("删除老婆").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			deleteWifeEvent, cancel := ctx.FutureEvent("message", ctx.CheckSession()).Repeat()
			defer cancel()

			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入要删除的角色名称喵~~~"))
			name := <-deleteWifeEvent
			rawName := strings.TrimSpace(name.Event.RawMessage)
			picPath := engine.DataFolder() + "wives/" + rawName + ".jpg"
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("正在删除老婆"+rawName+"嗒~~~"))

			cards, err := getWifeCards(engine.DataFolder() + "wife.json")
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("喵喵喵！老婆池加载失败了喵~~！", err),
				)
				return
			}
			for i := 0; i < len(cards); i++ {
				if cards[i].Name == rawName+".jpg" {
					cards = append(cards[:i], cards[i+1:]...)
					os.Remove(picPath)
					err = saveWifeFile(engine.DataFolder()+"wife.json", cards)
					if err != nil {
						logrus.Errorln(err)
						ctx.SendChain(
							message.At(ctx.Event.UserID),
							message.Text("保存老婆竟然出错了喵！！", err),
						)
						return
					}
					process.SleepAbout1sTo2s()
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("删除成功了喵~："))
					return
				}
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("没有找到要删除的老婆呢~"))
		})

}

func getWifeCards(path string) ([]wife, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cards []wife
	err = json.Unmarshal(data, &cards)
	if err != nil {
		return nil, err
	}
	return cards, nil
}

func saveWifeFile(path string, cards []wife) error {
	wifeJsonBytes, err := json.Marshal(cards)
	if err != nil {
		return err
	}
	return os.WriteFile(path, wifeJsonBytes, 0644)
}

func convertPictureToJpg(filePath string) error {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	img, err := png.Decode(bytes.NewReader(buf))
	if err != nil {
		img, err = jpeg.Decode(bytes.NewReader(buf))
		if err != nil {
			return err
		}
	}
	newBuf := bytes.Buffer{}
	err = jpeg.Encode(&newBuf, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return err
	}
	pos := strings.LastIndex(filePath, ".")
	outputPath := filePath[:pos] + ".jpg"
	err = os.WriteFile(outputPath, newBuf.Bytes(), 0644)
	if err == nil && filePath != outputPath {
		os.Remove(filePath)
	}
	return err
}

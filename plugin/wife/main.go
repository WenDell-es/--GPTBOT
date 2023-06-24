// Package wife 抽老婆
package wife

import (
	"encoding/json"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"os"
	"strings"
)

type wife struct {
	Name   string
	Source string
}

func init() {
	engine := control.Register("wife", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "- 抽老婆",
		Brief:            "从老婆库抽每日老婆",
		PublicDataFolder: "Wife",
	}).ApplySingle(ctxext.DefaultSingle)
	_ = os.MkdirAll(engine.DataFolder()+"wives", 0755)
	engine.OnFullMatch("抽老婆").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := os.ReadFile(engine.DataFolder() + "wife.json")
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("喵喵喵！老婆池加载失败了喵~~！", err),
				)
				return
			}
			var cards []wife
			err = json.Unmarshal(data, &cards)
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("喵喵喵！老婆池加载失败了喵~~！", err),
				)
				return
			}

			card := cards[fcext.RandSenderPerDayN(ctx.Event.UserID, len(cards))]
			data, err = os.ReadFile(engine.DataFolder() + "wives/" + card.Name)
			wifeName, _, _ := strings.Cut(card.Name, ".")
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("今天的二次元老婆是~【", wifeName, "】哒\n【图片下载失败: ", err, "】"),
				)
				return
			}
			if id := ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text("今天的二次元老婆是~【", wifeName, "】哒!\n来自作品【", card.Source, "】哦~"),
				message.ImageBytes(data),
			); id.ID() == 0 {
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("今天的二次元老婆是~【", wifeName, "】哒\n【图片发送失败, 请联系维护者】"),
				)
			}
		})
}

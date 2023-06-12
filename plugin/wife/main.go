// Package wife 抽老婆
package wife

import (
	"github.com/sirupsen/logrus"
	"os"
	"strings"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

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
			cards := []string{}
			dirs, err := os.ReadDir(engine.DataFolder() + "wives")
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("喵喵喵！老婆池加载失败了喵~~！", err),
				)
				return
			}
			for _, dir := range dirs {
				cards = append(cards, dir.Name())
			}
			card := cards[fcext.RandSenderPerDayN(ctx.Event.UserID, len(cards))]
			data, err := os.ReadFile(engine.DataFolder() + "wives/" + card)
			card, _, _ = strings.Cut(card, ".")
			if err != nil {
				logrus.Errorln(err)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("今天的二次元老婆是~【", card, "】哒\n【图片下载失败: ", err, "】"),
				)
				return
			}
			if id := ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text("今天的二次元老婆是~【", card, "】哒"),
				message.ImageBytes(data),
			); id.ID() == 0 {
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("今天的二次元老婆是~【", card, "】哒\n【图片发送失败, 请联系维护者】"),
				)
			}
		})
}

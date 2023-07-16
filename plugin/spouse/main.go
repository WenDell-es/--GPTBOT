package spouse

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	commandHandler "gptbot/plugin/spouse/handler"
	"gptbot/plugin/spouse/model"
)

const (
	HelpMessage = "每日抽配偶，每人每天只能抽一个（东八区0点更新）\n" +
		"以下为全部命令\n" +
		"抽老婆 (随机抽一个老婆)\n" +
		"抽老公 (随机抽一个老公)\n" +
		"添加老婆 （输入后按提示为老婆池添加老婆）\n" +
		"添加老公 （输入后按提示为老婆池添加老公）\n" +
		"删除老婆 （输入后按提示从老婆池删除老婆）\n" +
		"删除老公 （输入后按提示从老婆池删除老公）\n" +
		"老婆列表 (查看目前老婆池有哪些老婆)\n" +
		"老公列表 (查看目前老婆池有哪些老公)\n"
)

var engine *control.Engine

func init() {
	engine = control.Register("spouse", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Help:              HelpMessage,
		Brief:             "从池子中抽取每日配偶",
		PrivateDataFolder: "spouse",
	}).ApplySingle(ctxext.DefaultSingle)
	engine.OnFullMatch("添加老婆", zero.OnlyGroup, zero.MustProvidePicture).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			handler := commandHandler.NewAddSpouseHandler(engine.DataFolder(), model.Wife, ctx)
			if handler.CreateEventChan().FetchSpouseName().FetchSpouseSource().GetBaseCards().GetGroupCards().AddNewCard().DownloadPicture().
				ConvertPicture().UploadPictureToStore().UploadIndexFileToStore().Cancel().NotifyUser().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnFullMatch("删除老婆", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			handler := commandHandler.NewDeleteSpouseHandler(engine.DataFolder(), model.Wife, ctx)
			if handler.CreateEventChan().FetchSpouseName().GetBaseCards().GetGroupCards().DeleteCardIndex().DeletePicture().UploadIndexFileToStore().Cancel().
				NotifyUser().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnFullMatch("抽老婆", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			handler := commandHandler.NewRandomSpouseHandler(ctx, model.Wife)
			if handler.CheckRecords().GetBaseCards().GetGroupCards().FetchRandomCard().SendPicture().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnFullMatch("老婆列表", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			handler := commandHandler.NewListSpouseHandler(ctx, model.Wife)
			if handler.GetBaseCards().GetGroupCards().GenerateImageFont().SendImage().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnFullMatch("添加老公", zero.OnlyGroup, zero.MustProvidePicture).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			handler := commandHandler.NewAddSpouseHandler(engine.DataFolder(), model.Husband, ctx)
			if handler.CreateEventChan().FetchSpouseName().FetchSpouseSource().GetBaseCards().GetGroupCards().AddNewCard().DownloadPicture().
				ConvertPicture().UploadPictureToStore().UploadIndexFileToStore().Cancel().NotifyUser().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnFullMatch("删除老公", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			handler := commandHandler.NewDeleteSpouseHandler(engine.DataFolder(), model.Husband, ctx)
			if handler.CreateEventChan().FetchSpouseName().GetBaseCards().GetGroupCards().DeleteCardIndex().DeletePicture().UploadIndexFileToStore().Cancel().
				NotifyUser().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnFullMatch("抽老公", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			handler := commandHandler.NewRandomSpouseHandler(ctx, model.Husband)
			if handler.CheckRecords().GetBaseCards().GetGroupCards().FetchRandomCard().SendPicture().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnFullMatch("老公列表", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			handler := commandHandler.NewListSpouseHandler(ctx, model.Husband)
			if handler.GetBaseCards().GetGroupCards().GenerateImageFont().SendImage().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})
}

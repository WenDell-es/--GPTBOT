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
		"抽[老婆|老公] \n" +
		"添加[老婆|老公]\n" +
		"删除[老婆|老公]\n" +
		"更新[老婆|老公]\n" +
		"[老婆|老公]列表 (查看池子内容)\n" +
		"查看[老婆|老公] (查看某一个角色具体信息)\n"
)

var engine *control.Engine

func init() {
	engine = control.Register("spouse", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Help:              HelpMessage,
		Brief:             "从池子中抽取每日配偶",
		PrivateDataFolder: "spouse",
	}).ApplySingle(ctxext.DefaultSingle)
	engine.OnRegex(`^添加\s?(.*)$`, IsSupported, zero.OnlyGroup, zero.MustProvidePicture).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			spouseType := model.Type(keyword)
			handler := commandHandler.NewAddSpouseHandler(engine.DataFolder(), spouseType, ctx)
			if handler.CreateEventChan().FetchSpouseName().FetchSpouseSource().DownloadPicture().
				ConvertPicture().GetBaseCards().GetGroupCards().AddNewCard().UploadPictureToStore().UploadIndexFileToStore().Cancel().NotifyUser().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})
	// 管理员命令，向公共卡池中添加spouse
	engine.OnRegex(`^ab\s?(.*)$`, zero.SuperUserPermission, zero.MustProvidePicture, IsSupported).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			spouseType := model.Type(keyword)
			handler := commandHandler.NewAddSpouseHandler(engine.DataFolder(), spouseType, ctx)
			if handler.CreateEventChan().FetchSpouseName().FetchSpouseSource().SetBaseMode().DownloadPicture().ConvertPicture().GetGroupCards().AddNewCard().
				UploadPictureToStore().UploadIndexFileToStore().Cancel().NotifyUser().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnRegex(`^删除\s?(.*)$`, zero.OnlyGroup, IsSupported).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			spouseType := model.Type(keyword)
			handler := commandHandler.NewDeleteSpouseHandler(engine.DataFolder(), spouseType, ctx)
			if handler.CreateEventChan().FetchSpouseName().GetBaseCards().GetGroupCards().DeleteCardIndex().DeletePicture().UploadIndexFileToStore().Cancel().
				NotifyUser().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnRegex(`^抽\s?(.*)$`, zero.OnlyGroup, IsSupported).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			spouseType := model.Type(keyword)
			handler := commandHandler.NewRandomSpouseHandler(ctx, spouseType)
			if handler.CheckRecords().GetBaseCards().GetGroupCards().GetGroupWeights().FetchRandomCard().UploadWeightFileToStore().SendPicture().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnRegex(`^(.*)列表$`, zero.OnlyGroup, IsSupported).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			spouseType := model.Type(keyword)
			handler := commandHandler.NewListSpouseHandler(ctx, spouseType)
			if handler.GetBaseCards().GetGroupCards().GenerateImageFont().SendImage().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnRegex(`^查看\s?(.*)$`, zero.OnlyGroup, IsSupported).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			spouseType := model.Type(keyword)
			handler := commandHandler.NewQuerySpouseHandler(ctx, spouseType)
			if handler.CreateEventChan().FetchSpouseName().GetBaseCards().GetGroupCards().GetGroupWeights().CheckCards().SendPicture().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})

	engine.OnRegex(`^更新\s?(.*)$`, zero.OnlyGroup, IsSupported, zero.MustProvidePicture).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			spouseType := model.Type(keyword)
			handler := commandHandler.NewUpdateSpouseHandler(engine.DataFolder(), spouseType, ctx)
			if handler.CreateEventChan().FetchSpouseName().DownloadPicture().ConvertPicture().GetBaseCards().GetGroupCards().UpdateCard().
				UploadPictureToStore().UploadIndexFileToStore().Cancel().NotifyUser().Err() != nil {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(handler.Err().Error()))
				logrus.Errorln(handler.Err().Error())
			}
		})
}

func IsSupported(ctx *zero.Ctx) bool {
	keyword := ctx.State["regex_matched"].([]string)[1]
	spouseType := model.Type(keyword)
	if spouseType.String() == "" {
		return false
	}
	return true
}

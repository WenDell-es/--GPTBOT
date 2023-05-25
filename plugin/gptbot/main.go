package gptbot

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/go-ini/ini"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gptbot/plugin/gptbot/botservice"
	_ "gptbot/plugin/gptbot/chatgpt"
	"gptbot/plugin/gptbot/config"
	"gptbot/plugin/gptbot/constants"
	"gptbot/plugin/gptbot/model"
	"gptbot/plugin/gptbot/util"
	"strconv"
)

func init() {
	engine := control.Register("gptbot", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "gpt机器人",
		Help:              constants.HelpContent,
		PrivateDataFolder: "gptbot",
	})

	conf, err := ini.Load(engine.DataFolder() + "conf.ini")
	if err != nil {
		logrus.Fatalln("加载gpt机器人配置错误", err)
	}
	cfg := config.ChatGptConfig{}
	if err = conf.MapTo(&cfg); err != nil {
		logrus.Fatalln("解析gpt机器人配置错误", err)
	}
	gptBot := botservice.NewGptBot(cfg)

	engine.OnCommand("查看提示词").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("当前提示词为：\n\n" + gptBot.GetChat(util.GetChatId(ctx)).GetPrompt().Content))
	})
	engine.OnCommand("设置提示词").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		prompt := ctx.State["args"].(string)
		gptBot.GetChat(util.GetChatId(ctx)).SetPrompt(prompt)
		ctx.SendChain(message.Text("设置提示词成功，当前提示词为：\n\n" + gptBot.GetChat(util.GetChatId(ctx)).GetPrompt().Content))
	})
	engine.OnCommand("查看gpt模型").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("当前gpt模型为：\n\n" + gptBot.GetChat(util.GetChatId(ctx)).GetModel()))
	})
	engine.OnCommand("设置gpt模型").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		m := ctx.State["args"].(string)
		if err := gptBot.GetChat(util.GetChatId(ctx)).SetModel(m); err != nil {
			ctx.SendChain(message.Text(errors.Wrap(err, "设置gpt模型错误").Error()))
			return
		}
		ctx.SendChain(message.Text("设置gpt模型成功，当前gpt模型为：\n\n" + gptBot.GetChat(util.GetChatId(ctx)).GetModel()))
	})
	engine.OnCommand("查看记忆区").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		messages := gptBot.GetChat(util.GetChatId(ctx)).GetMessages()
		resp := "记忆区：\n"
		for _, m := range messages {
			resp += m.Name + ":" + m.Content + "\n"
		}
		ctx.SendChain(message.Text(resp))
	})
	engine.OnCommand("清空记忆区").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gptBot.GetChat(util.GetChatId(ctx)).ClearMessages()
		ctx.SendChain(message.Text("已清空记忆区"))

	})
	engine.OnCommand("查看群回复概率").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("当前群回复概率：", gptBot.GetChat(util.GetChatId(ctx)).GetGroupProbability()))
	})
	engine.OnCommand("设置群回复概率").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		probStr := ctx.State["args"].(string)
		prob, err := strconv.Atoi(probStr)
		if err != nil {
			ctx.SendChain(message.Text(errors.Wrap(err, "概率无法解析成int").Error()))
			return
		}
		gptBot.GetChat(util.GetChatId(ctx)).SetGroupProbability(prob)
		ctx.SendChain(message.Text("设置成功!\n当前群回复概率：", gptBot.GetChat(util.GetChatId(ctx)).GetGroupProbability()))

	})

	matcher := engine.OnMessage(zero.OnlyToMe, zero.OnlyGroup, func(ctx *zero.Ctx) bool {
		return !zero.HasPicture(ctx)
	}).SetBlock(true).Limit(ctxext.LimitByGroup)
	(*zero.Matcher)(matcher).SetPriority(matcher.Priority).Handle(func(ctx *zero.Ctx) {
		gptBot.GetChat(util.GetChatId(ctx)).AddMessage(&model.Message{
			Role:    "user",
			Content: ctx.Event.Message.String(),
			Name:    strconv.FormatInt(ctx.Event.UserID, 10),
		})
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(gptBot.Talk(ctx)))
	})

	matcher = engine.OnMessage(zero.OnlyGroup, func(ctx *zero.Ctx) bool {
		return !zero.HasPicture(ctx)
	}).SetBlock(true).Limit(ctxext.LimitByGroup)
	(*zero.Matcher)(matcher).SetPriority(matcher.Priority + 1).Handle(func(ctx *zero.Ctx) {
		currentChat := gptBot.GetChat(util.GetChatId(ctx))
		currentChat.AddMessage(&model.Message{
			Role:    "user",
			Content: ctx.Event.Message.String(),
			Name:    strconv.FormatInt(ctx.Event.UserID, 10),
		})
		if !currentChat.GroupChatCheck() {
			return
		}
		ctx.SendChain(message.Text(gptBot.Talk(ctx)))
	})

	matcher = engine.OnMessage(zero.OnlyPrivate, func(ctx *zero.Ctx) bool {
		return !zero.HasPicture(ctx)
	}).SetBlock(true).Limit(ctxext.LimitByUser)
	(*zero.Matcher)(matcher).SetPriority(matcher.Priority + 1).Handle(func(ctx *zero.Ctx) {
		gptBot.GetChat(util.GetChatId(ctx)).AddMessage(&model.Message{
			Role:    "user",
			Content: ctx.Event.Message.String(),
			Name:    strconv.FormatInt(ctx.Event.UserID, 10),
		})
		ctx.SendChain(message.Text(gptBot.Talk(ctx)))
	})
}

package handler

import (
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	imageFont "gptbot/imageFont"
	"gptbot/plugin/spouse/model"
	"gptbot/plugin/spouse/util"
)

type ListSpouseHandler struct {
	mainCtx    *zero.Ctx
	spouseType model.Type
	err        error
	listSpouseInternal
}

type listSpouseInternal struct {
	groupId    int64
	baseCards  []model.Card
	groupCards []model.Card
	imgFont    *imageFont.ImageFont
}

func NewListSpouseHandler(ctx *zero.Ctx, spouseType model.Type) *ListSpouseHandler {
	return &ListSpouseHandler{
		mainCtx:    ctx,
		spouseType: spouseType,
	}
}
func (h *ListSpouseHandler) Err() error {
	return h.err
}

func (h *ListSpouseHandler) GetBaseCards() *ListSpouseHandler {
	if h.err != nil {
		return h
	}
	h.baseCards, h.err = util.GetCards(int64(0), h.spouseType)
	return h
}

func (h *ListSpouseHandler) GetGroupCards() *ListSpouseHandler {
	if h.err != nil {
		return h
	}
	h.groupCards, h.err = util.GetCards(h.mainCtx.Event.GroupID, h.spouseType)
	return h
}

func (h *ListSpouseHandler) GenerateImageFont() *ListSpouseHandler {
	if h.err != nil {
		return h
	}
	h.imgFont, h.err = imageFont.NewImageFont()
	return h
}

func (h *ListSpouseHandler) SendImage() *ListSpouseHandler {
	if h.err != nil {
		return h
	}
	cards := append(h.baseCards, h.groupCards...)
	h.mainCtx.SendChain(message.Text("当前共有", len(cards), "位"+h.spouseType.String()+"哦~"))
	h.mainCtx.SendChain(message.Text("正在获取" + h.spouseType.String() + "名单，请稍候喵~"))
	msg := message.Message{}
	var num int
	for i := 0; i < len(cards); i++ {
		h.imgFont.Write(cards[i].Name)
		num++
		if num >= 22 {
			msg = append(msg, ctxext.FakeSenderForwardNode(h.mainCtx, message.ImageBytes(h.imgFont.GetImage())))
			h.imgFont.Clear()
			num = 0
		}
	}
	if num > 0 {
		msg = append(msg, ctxext.FakeSenderForwardNode(h.mainCtx, message.ImageBytes(h.imgFont.GetImage())))
	}
	h.mainCtx.Send(msg)
	return h
}

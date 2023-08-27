package handler

import (
	"github.com/pkg/errors"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gptbot/imageFont"
	"gptbot/plugin/spouse/model"
	"gptbot/plugin/spouse/util"
	"gptbot/store"
	"strings"
)

type QuerySpouseHandler struct {
	mainCtx    *zero.Ctx
	spouseType model.Type
	name       string
	err        error
	querySpouseInternal
}

type querySpouseInternal struct {
	groupId    int64
	baseCards  []model.Card
	groupCards []model.Card
	imgFont    *imageFont.ImageFont
	card       model.Card
	event      <-chan *zero.Ctx
	cancel     func()
}

func NewQuerySpouseHandler(ctx *zero.Ctx, spouseType model.Type) *QuerySpouseHandler {
	return &QuerySpouseHandler{
		mainCtx:    ctx,
		spouseType: spouseType,
	}
}

func (h *QuerySpouseHandler) Err() error {
	return h.err
}

func (h *QuerySpouseHandler) CreateEventChan() *QuerySpouseHandler {
	h.event, h.cancel = h.mainCtx.FutureEvent("message", h.mainCtx.CheckSession()).Repeat()
	return h
}

func (h *QuerySpouseHandler) FetchSpouseName() *QuerySpouseHandler {
	if h.err != nil {
		return h
	}
	h.mainCtx.SendChain(message.At(h.mainCtx.Event.UserID), message.Text("请输入要查询的角色名称喵~~~"))
	name, err := getUserInput(h.event)
	h.err = err
	h.name = strings.TrimSpace(name)
	h.cancel()
	return h
}

func (h *QuerySpouseHandler) GetBaseCards() *QuerySpouseHandler {
	h.baseCards, h.err = util.GetCards(int64(0), h.spouseType)
	return h
}

func (h *QuerySpouseHandler) GetGroupCards() *QuerySpouseHandler {
	if h.err != nil {
		return h
	}
	h.groupCards, h.err = util.GetCards(h.mainCtx.Event.GroupID, h.spouseType)
	return h
}
func (h *QuerySpouseHandler) CheckCards() *QuerySpouseHandler {
	if h.err != nil {
		return h
	}
	cards := append(h.baseCards, h.groupCards...)

	for _, card := range cards {
		if card.Name == h.name {
			h.card = card
			return h
		}
	}
	h.err = errors.New("没有找到" + h.name)
	return h
}

func (h *QuerySpouseHandler) SendPicture() *QuerySpouseHandler {
	if h.err != nil {
		return h
	}
	url := store.GetStoreClient().GetObjectUrl(util.GetPicturePath(h.card.GroupId, h.spouseType) + h.card.Hash + ".jpg")
	if id := h.mainCtx.SendChain(
		message.At(h.mainCtx.Event.UserID),
		message.Text("类别：", h.spouseType.String(), "\n"+
			"名称：", h.card.Name, "\n"+
			"作品名：", h.card.Source, "\n"+
			"上传人昵称：", h.card.UploaderName, "\n"+
			"上传人QQ号：", h.card.UploaderId, "\n"+
			"卡池编号：", h.card.GroupId),
		message.Image(url),
	); id.ID() == 0 {
		h.mainCtx.SendChain(
			message.At(h.mainCtx.Event.UserID),
			message.Text("【图片发送失败,被腾讯风控系统拦截。 请联系维护者】"),
		)
	}
	return h
}

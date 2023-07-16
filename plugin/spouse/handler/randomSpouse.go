package handler

import (
	fcext "github.com/FloatTech/floatbox/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gptbot/plugin/spouse/model"
	"gptbot/plugin/spouse/records"
	"gptbot/plugin/spouse/util"
	"gptbot/store"
	"time"
)

type RandomSpouseHandler struct {
	mainCtx    *zero.Ctx
	spouseType model.Type
	err        error
	randomSpouseInternal
}

type randomSpouseInternal struct {
	groupId    int64
	card       *model.Card
	baseCards  []model.Card
	groupCards []model.Card
}

func (h *RandomSpouseHandler) Err() error {
	return h.err
}

func NewRandomSpouseHandler(mainCtx *zero.Ctx, spouseType model.Type) *RandomSpouseHandler {
	return &RandomSpouseHandler{
		mainCtx:    mainCtx,
		spouseType: spouseType,
	}
}

func (h *RandomSpouseHandler) CheckRecords() *RandomSpouseHandler {
	if rd, ok := records.UserRecords.Load(h.mainCtx.Event.UserID); ok &&
		rd.(map[model.Type]records.Record)[h.spouseType].Date.Format("20060102") == time.Now().Format("20060102") {
		cd := rd.(map[model.Type]records.Record)[h.spouseType].Card
		h.card = &cd
	}
	return h
}

func (h *RandomSpouseHandler) GetBaseCards() *RandomSpouseHandler {
	if h.err != nil || h.card != nil {
		return h
	}
	h.baseCards, h.err = util.GetCards(int64(0), h.spouseType)
	return h
}

func (h *RandomSpouseHandler) GetGroupCards() *RandomSpouseHandler {
	if h.err != nil || h.card != nil {
		return h
	}
	h.groupCards, h.err = util.GetCards(h.mainCtx.Event.GroupID, h.spouseType)
	return h
}
func (h *RandomSpouseHandler) FetchRandomCard() *RandomSpouseHandler {
	if h.err != nil || h.card != nil {
		return h
	}
	cards := append(h.baseCards, h.groupCards...)
	card := cards[fcext.RandSenderPerDayN(h.mainCtx.Event.UserID, len(cards))]
	h.card = &card

	rdMap := make(map[model.Type]records.Record)
	if rd, ok := records.UserRecords.Load(h.mainCtx.Event.UserID); ok {
		rdMap = rd.(map[model.Type]records.Record)
	}
	rdMap[h.spouseType] = records.Record{
		Date: time.Now(),
		Card: card,
	}
	records.UserRecords.Store(h.mainCtx.Event.UserID, rdMap)
	return h
}

func (h *RandomSpouseHandler) SendPicture() *RandomSpouseHandler {
	if h.err != nil {
		return h
	}
	url := store.GetStoreClient().GetObjectUrl(util.GetPicturePath(h.card.GroupId, h.spouseType) + h.card.Name + ".jpg")
	if id := h.mainCtx.SendChain(
		message.At(h.mainCtx.Event.UserID),
		message.Text("今天的二次元"+h.spouseType.String()+"是~【", h.card.Name, "】哒!\n来自作品【", h.card.Source, "】哦~\n上传人是【", h.card.UploaderName, ",", h.card.UploaderId, "】呢"),
		message.Image(url),
	); id.ID() == 0 {
		h.mainCtx.SendChain(
			message.At(h.mainCtx.Event.UserID),
			message.Text("今天的二次元"+h.spouseType.String()+"是~【", h.card.Name, "】哒\n【图片发送失败, 请联系维护者】"),
		)
	}
	return h
}

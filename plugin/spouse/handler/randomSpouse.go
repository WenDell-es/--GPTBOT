package handler

import (
	"encoding/json"
	"github.com/tencentyun/cos-go-sdk-v5"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gptbot/plugin/spouse/model"
	"gptbot/plugin/spouse/random"
	"gptbot/plugin/spouse/records"
	"gptbot/plugin/spouse/util"
	"gptbot/store"
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
	weights    map[string]float64
	hasRecord  bool
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
	if records.GetSpouseRecorder().HasSpouseToday(h.mainCtx.Event.UserID, h.mainCtx.Event.GroupID, h.spouseType) {
		h.card = records.GetSpouseRecorder().GetSpouseToday(h.mainCtx.Event.UserID, h.mainCtx.Event.GroupID, h.spouseType)
		h.hasRecord = true
	}
	return h
}

func (h *RandomSpouseHandler) GetBaseCards() *RandomSpouseHandler {
	if h.err != nil || h.hasRecord {
		return h
	}
	h.baseCards, h.err = util.GetCards(int64(0), h.spouseType)
	return h
}

func (h *RandomSpouseHandler) GetGroupCards() *RandomSpouseHandler {
	if h.err != nil || h.hasRecord {
		return h
	}
	h.groupCards, h.err = util.GetCards(h.mainCtx.Event.GroupID, h.spouseType)
	return h
}

func (h *RandomSpouseHandler) GetGroupWeights() *RandomSpouseHandler {
	if h.err != nil || h.hasRecord {
		return h
	}

	buf, err := store.GetStoreClient().GetObjectBytes(util.GetWeightPath(h.mainCtx.Event.GroupID, h.spouseType))
	if err != nil && !cos.IsNotFoundError(err) {
		h.err = err
		return h
	}
	weight := make(map[string]float64)
	_ = json.Unmarshal(buf, &weight)
	h.weights = weight
	return h
}
func (h *RandomSpouseHandler) FetchRandomCard() *RandomSpouseHandler {
	if h.err != nil || h.hasRecord {
		return h
	}
	cards := append(h.baseCards, h.groupCards...)
	card := random.GetRandomCard(cards, h.weights)
	h.card = &card
	records.GetSpouseRecorder().AddSpouseToday(h.mainCtx.Event.UserID, h.mainCtx.Event.GroupID, h.spouseType, &card)
	return h
}

func (h *RandomSpouseHandler) UploadWeightFileToStore() *RandomSpouseHandler {
	if h.err != nil || h.hasRecord {
		return h
	}
	weightJsonBytes, _ := json.Marshal(h.weights)
	h.err = store.GetStoreClient().UploadObjectByBytes(weightJsonBytes, util.GetWeightPath(h.mainCtx.Event.GroupID, h.spouseType))
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

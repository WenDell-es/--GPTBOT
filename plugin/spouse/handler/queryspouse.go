package handler

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tencentyun/cos-go-sdk-v5"
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
	probability float64
	weights     map[string]float64
	groupId     int64
	baseCards   []model.Card
	groupCards  []model.Card
	imgFont     *imageFont.ImageFont
	card        *model.Card
	event       <-chan *zero.Ctx
	cancel      func()
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

func (h *QuerySpouseHandler) GetGroupWeights() *QuerySpouseHandler {
	if h.err != nil {
		return h
	}
	buf, err := store.GetStoreClient().GetObjectBytes(util.GetWeightPath(h.mainCtx.Event.GroupID, h.spouseType))
	if err != nil && !cos.IsNotFoundError(err) {
		h.err = err
		return h
	}
	weight := make(map[string]float64)
	h.err = json.Unmarshal(buf, &weight)
	h.weights = weight
	return h
}

func (h *QuerySpouseHandler) CheckCards() *QuerySpouseHandler {
	if h.err != nil {
		return h
	}
	cards := append(h.baseCards, h.groupCards...)
	total := 0.0
	newWeight := 0.0
	for i := 0; i < len(cards); i++ {
		if _, ok := h.weights[cards[i].Name]; !ok {
			continue
		}
		total += h.weights[cards[i].Name]
	}
	newWeight = total/float64(len(cards)) + 1
	total = 0.0
	for i := 0; i < len(cards); i++ {
		if _, ok := h.weights[cards[i].Name]; !ok {
			h.weights[cards[i].Name] = newWeight
		}
		total += h.weights[cards[i].Name]
	}

	target := model.Card{}
	for _, card := range cards {
		if card.Name == h.name {
			target = card
			break
		}
	}
	h.card = &target
	if h.card.Name == "" {
		h.err = errors.New("没有找到" + h.name)
		return h
	}
	h.probability = h.weights[h.card.Name] / total
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
			"卡池编号：", h.card.GroupId, "\n"+
			"当前抽中概率：", fmt.Sprintf("%.6f", h.probability*100), "%"),
		message.Image(url),
	); id.ID() == 0 {
		h.mainCtx.SendChain(
			message.At(h.mainCtx.Event.UserID),
			message.Text("【图片发送失败,被腾讯风控系统拦截。 请联系维护者】"),
		)
	}
	return h
}

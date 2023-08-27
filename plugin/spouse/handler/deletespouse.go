package handler

import (
	"encoding/json"
	"errors"
	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gptbot/plugin/spouse/model"
	"gptbot/plugin/spouse/util"
	"gptbot/store"
	"os"
	"strconv"
	"strings"
)

type DeleteSpouseHandler struct {
	mainCtx    *zero.Ctx
	basePath   string
	spouseType model.Type
	err        error
	deleteSpouseInternal
}

type deleteSpouseInternal struct {
	name       string
	hash       string
	groupId    int64
	baseCards  []model.Card
	groupCards []model.Card
	event      <-chan *zero.Ctx
	cancel     func()
}

func NewDeleteSpouseHandler(basePath string, spouseType model.Type, mainCtx *zero.Ctx) *DeleteSpouseHandler {
	return &DeleteSpouseHandler{basePath: basePath, spouseType: spouseType, mainCtx: mainCtx}
}

func (h *DeleteSpouseHandler) Err() error {
	return h.err
}

func (h *DeleteSpouseHandler) CreateEventChan() *DeleteSpouseHandler {
	h.event, h.cancel = h.mainCtx.FutureEvent("message", h.mainCtx.CheckSession()).Repeat()
	return h
}

func (h *DeleteSpouseHandler) FetchSpouseName() *DeleteSpouseHandler {
	if h.err != nil {
		return h
	}
	h.mainCtx.SendChain(message.At(h.mainCtx.Event.UserID), message.Text("请输入要删除的角色名称喵~~~"))
	name, err := getUserInput(h.event)
	h.err = err
	h.name = strings.TrimSpace(name)
	return h
}

func (h *DeleteSpouseHandler) GetBaseCards() *DeleteSpouseHandler {
	if h.err != nil {
		return h
	}
	h.baseCards, h.err = util.GetCards(int64(0), h.spouseType)
	return h
}

func (h *DeleteSpouseHandler) GetGroupCards() *DeleteSpouseHandler {
	if h.err != nil {
		return h
	}
	h.groupCards, h.err = util.GetCards(h.mainCtx.Event.GroupID, h.spouseType)
	return h
}

func (h *DeleteSpouseHandler) DeleteCardIndex() *DeleteSpouseHandler {
	if h.err != nil {
		return h
	}

	for i := 0; i < len(h.groupCards); i++ {
		if h.groupCards[i].Name == h.name {
			h.hash = h.groupCards[i].Hash
			h.groupCards = append(h.groupCards[:i], h.groupCards[i+1:]...)
			h.groupId = h.mainCtx.Event.GroupID
			return h
		}
	}

	for i := 0; i < len(h.baseCards); i++ {
		if h.baseCards[i].Name == h.name {
			if zero.SuperUserPermission(h.mainCtx) {
				h.hash = h.baseCards[i].Hash
				h.baseCards = append(h.baseCards[:i], h.baseCards[i+1:]...)
				h.groupId = int64(0)
				return h
			}
			h.err = errors.New("只有Master大人才能删除这个角色哦~。如果你想删除这个角色，请联系机器人的管理员哦")
			return h
		}
	}

	h.err = errors.New("没有找到这个角色~TNT~")
	return h
}

func (h *DeleteSpouseHandler) DeletePicture() *DeleteSpouseHandler {
	if h.err != nil {
		return h
	}
	h.err = store.GetStoreClient().DeleteObject(util.GetPicturePath(h.groupId, h.spouseType) + h.hash + ".jpg")
	return h
}

func (h *DeleteSpouseHandler) UploadIndexFileToStore() *DeleteSpouseHandler {
	if h.err != nil {
		return h
	}
	var cards []model.Card
	if h.groupId == h.mainCtx.Event.GroupID {
		cards = h.groupCards
	} else {
		cards = h.baseCards
	}

	tempath, _ := os.MkdirTemp(h.basePath, strconv.FormatInt(h.mainCtx.Event.GroupID, 10))
	tempath += "/"

	wifeJsonBytes, _ := json.MarshalIndent(cards, "", "  ")
	_ = os.WriteFile(tempath+"index.json", wifeJsonBytes, 0644)

	h.err = store.GetStoreClient().UploadObject(tempath+"index.json", util.GetIndexPath(h.groupId, h.spouseType))
	_ = os.RemoveAll(tempath)
	return h
}

func (h *DeleteSpouseHandler) NotifyUser() *DeleteSpouseHandler {
	if h.err != nil {
		return h
	}
	process.SleepAbout1sTo2s()
	h.mainCtx.SendChain(message.Reply(h.mainCtx.Event.MessageID), message.Text("删除成功了喵！"))
	return h
}

func (h *DeleteSpouseHandler) Cancel() *DeleteSpouseHandler {
	if h.cancel != nil {
		h.cancel()
	}
	return h
}

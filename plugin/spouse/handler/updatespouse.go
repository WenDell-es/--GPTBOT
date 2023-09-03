package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/process"
	"github.com/pkg/errors"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gptbot/plugin/spouse/model"
	"gptbot/plugin/spouse/util"
	"gptbot/store"
	"os"
	"strconv"
	"strings"
)

type UpdateSpouseHandler struct {
	mainCtx    *zero.Ctx
	basePath   string
	spouseType model.Type
	err        error
	updateSpouseInternal
}

type updateSpouseInternal struct {
	card       model.Card
	baseCards  []model.Card
	groupCards []model.Card
	event      <-chan *zero.Ctx
	cancel     func()
	groupPath  string
	gid        int64
	oldHash    string
}

func NewUpdateSpouseHandler(basePath string, spouseType model.Type, mainCtx *zero.Ctx) *UpdateSpouseHandler {
	return &UpdateSpouseHandler{basePath: basePath, spouseType: spouseType, mainCtx: mainCtx}
}

func (h *UpdateSpouseHandler) Err() error {
	return h.err
}

func (h *UpdateSpouseHandler) CreateEventChan() *UpdateSpouseHandler {
	h.event, h.cancel = h.mainCtx.FutureEvent("message", h.mainCtx.CheckSession()).Repeat()
	h.gid = h.mainCtx.Event.GroupID
	return h
}

func (h *UpdateSpouseHandler) FetchSpouseName() *UpdateSpouseHandler {
	if h.err != nil {
		return h
	}
	h.mainCtx.SendChain(message.At(h.mainCtx.Event.UserID), message.Text("请输入要更新"+h.spouseType.String()+"的名称喵~~"))
	name, err := getUserInput(h.event)
	h.err = err
	h.card.Name = strings.TrimSpace(name)
	return h
}

func (h *UpdateSpouseHandler) DownloadPicture() *UpdateSpouseHandler {
	if h.err != nil {
		return h
	}
	url := h.mainCtx.State["image_url"].([]string)[0]

	gp, err := os.MkdirTemp(h.basePath, strconv.FormatInt(h.gid, 10))
	if err != nil {
		h.err = err
		return h
	}
	h.groupPath = gp + "/"
	h.err = file.DownloadTo(url, h.groupPath+h.card.Name+".jpg")
	return h
}

func (h *UpdateSpouseHandler) ConvertPicture() *UpdateSpouseHandler {
	if h.err != nil {
		return h
	}
	h.err = convertPictureToJpg(h.groupPath + h.card.Name + ".jpg")
	buf, err := os.ReadFile(h.groupPath + h.card.Name + ".jpg")
	if err != nil {
		h.err = err
		return h
	}
	sum := md5.Sum(buf)
	h.card.Hash = hex.EncodeToString(sum[:])
	return h
}

func (h *UpdateSpouseHandler) GetBaseCards() *UpdateSpouseHandler {
	if h.err != nil {
		return h
	}
	h.baseCards, h.err = util.GetCards(int64(0), h.spouseType)
	return h
}

func (h *UpdateSpouseHandler) GetGroupCards() *UpdateSpouseHandler {
	if h.err != nil {
		return h
	}
	h.groupCards, h.err = util.GetCards(h.gid, h.spouseType)
	return h
}

func (h *UpdateSpouseHandler) UpdateCard() *UpdateSpouseHandler {
	if h.err != nil {
		return h
	}
	for i := 0; i < len(h.groupCards); i++ {
		if h.groupCards[i].Name == h.card.Name {
			h.oldHash = h.groupCards[i].Hash
			h.groupCards[i].Hash = h.card.Hash
			return h
		}
	}
	for i := 0; i < len(h.baseCards); i++ {
		if h.baseCards[i].Name == h.card.Name {
			if zero.SuperUserPermission(h.mainCtx) {
				h.oldHash = h.baseCards[i].Hash
				h.baseCards[i].Hash = h.card.Hash
				h.gid = 0
				return h
			}
			h.err = errors.New("只有Master大人才能更改这个角色哦~。如果你想更改这个角色，请联系机器人的管理员哦")
			return h
		}
	}
	h.err = errors.New("没有找到这个角色~TNT~")
	return h
}

func (h *UpdateSpouseHandler) UploadPictureToStore() *UpdateSpouseHandler {
	if h.err != nil {
		return h
	}
	h.err = store.GetStoreClient().UploadObject(h.groupPath+h.card.Name+".jpg", util.GetPicturePath(h.gid, h.spouseType)+h.card.Hash+".jpg")
	h.err = store.GetStoreClient().DeleteObject(util.GetPicturePath(h.gid, h.spouseType) + h.oldHash + ".jpg")
	return h
}

func (h *UpdateSpouseHandler) UploadIndexFileToStore() *UpdateSpouseHandler {
	if h.err != nil {
		return h
	}
	cards := h.groupCards
	if h.gid == 0 {
		cards = h.baseCards
	}
	wifeJsonBytes, _ := json.MarshalIndent(cards, "", "  ")

	_ = os.WriteFile(h.groupPath+"index.json", wifeJsonBytes, 0644)

	h.err = store.GetStoreClient().UploadObject(h.groupPath+"index.json", util.GetIndexPath(h.gid, h.spouseType))
	_ = os.RemoveAll(h.groupPath)
	return h
}

func (h *UpdateSpouseHandler) NotifyUser() *UpdateSpouseHandler {
	if h.err != nil {
		_ = os.RemoveAll(h.groupPath)
		return h
	}
	process.SleepAbout1sTo2s()
	h.mainCtx.SendChain(message.Reply(h.mainCtx.Event.MessageID), message.Text("成功！"))
	return h
}

func (h *UpdateSpouseHandler) Cancel() *UpdateSpouseHandler {
	if h.cancel != nil {
		h.cancel()
	}
	return h
}

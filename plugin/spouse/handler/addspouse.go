package handler

import (
	"bytes"
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
	"image/jpeg"
	"image/png"
	"os"
	"strconv"
	"strings"
	"time"
)

type AddSpouseHandler struct {
	mainCtx    *zero.Ctx
	basePath   string
	spouseType model.Type
	err        error
	internal
}

type internal struct {
	card       model.Card
	baseCards  []model.Card
	groupCards []model.Card
	event      <-chan *zero.Ctx
	cancel     func()
	groupPath  string
	gid        int64
}

func NewAddSpouseHandler(basePath string, spouseType model.Type, mainCtx *zero.Ctx) *AddSpouseHandler {
	return &AddSpouseHandler{basePath: basePath, spouseType: spouseType, mainCtx: mainCtx}
}

func (h *AddSpouseHandler) Err() error {
	return h.err
}

func (h *AddSpouseHandler) CreateEventChan() *AddSpouseHandler {
	h.event, h.cancel = h.mainCtx.FutureEvent("message", h.mainCtx.CheckSession()).Repeat()
	h.gid = h.mainCtx.Event.GroupID
	return h
}

func (h *AddSpouseHandler) FetchSpouseName() *AddSpouseHandler {
	if h.err != nil {
		return h
	}
	h.mainCtx.SendChain(message.At(h.mainCtx.Event.UserID), message.Text("请输入新"+h.spouseType.String()+"的名称喵~~"))
	name, err := getUserInput(h.event)
	h.err = err
	h.card.Name = strings.TrimSpace(name)
	return h
}

func (h *AddSpouseHandler) FetchSpouseSource() *AddSpouseHandler {
	if h.err != nil {
		return h
	}
	h.mainCtx.SendChain(message.At(h.mainCtx.Event.UserID), message.Text("接下来请为"+h.card.Name+"添加角色出处哦~"))
	source, err := getUserInput(h.event)
	h.err = err
	h.card.Source = strings.TrimSpace(source)
	h.card.UploaderId = h.mainCtx.Event.UserID
	h.card.UploaderName = h.mainCtx.Event.Sender.NickName
	h.card.GroupId = h.gid
	return h
}

func (h *AddSpouseHandler) SetBaseMode() *AddSpouseHandler {
	if h.err != nil {
		return h
	}
	h.card.UploaderId = 0
	h.card.UploaderName = "system"
	h.card.GroupId = 0
	h.gid = 0
	return h
}

func (h *AddSpouseHandler) DownloadPicture() *AddSpouseHandler {
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

func (h *AddSpouseHandler) ConvertPicture() *AddSpouseHandler {
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

func (h *AddSpouseHandler) GetBaseCards() *AddSpouseHandler {
	if h.err != nil {
		return h
	}
	h.baseCards, h.err = util.GetCards(int64(0), h.spouseType)
	return h
}

func (h *AddSpouseHandler) GetGroupCards() *AddSpouseHandler {
	if h.err != nil {
		return h
	}
	h.groupCards, h.err = util.GetCards(h.gid, h.spouseType)
	return h
}

func (h *AddSpouseHandler) AddNewCard() *AddSpouseHandler {
	if h.err != nil {
		return h
	}
	allCards := append(h.baseCards, h.groupCards...)
	for _, card := range allCards {
		if card.Name == h.card.Name {
			h.err = errors.New(h.spouseType.String() + h.card.Name + "已经存在啦!")
			return h
		}
	}
	h.groupCards = append(h.groupCards, h.card)
	return h
}

func (h *AddSpouseHandler) UploadPictureToStore() *AddSpouseHandler {
	if h.err != nil {
		return h
	}
	h.err = store.GetStoreClient().UploadObject(h.groupPath+h.card.Name+".jpg", util.GetPicturePath(h.gid, h.spouseType)+h.card.Hash+".jpg")
	return h
}

func (h *AddSpouseHandler) UploadIndexFileToStore() *AddSpouseHandler {
	if h.err != nil {
		return h
	}
	wifeJsonBytes, _ := json.MarshalIndent(h.groupCards, "", "  ")

	_ = os.WriteFile(h.groupPath+"index.json", wifeJsonBytes, 0644)

	h.err = store.GetStoreClient().UploadObject(h.groupPath+"index.json", util.GetIndexPath(h.gid, h.spouseType))
	_ = os.RemoveAll(h.groupPath)
	return h
}

func (h *AddSpouseHandler) NotifyUser() *AddSpouseHandler {
	if h.err != nil {
		_ = os.RemoveAll(h.groupPath)
		return h
	}
	process.SleepAbout1sTo2s()
	h.mainCtx.SendChain(message.Reply(h.mainCtx.Event.MessageID), message.Text("成功！"))
	return h
}

func (h *AddSpouseHandler) Cancel() *AddSpouseHandler {
	if h.cancel != nil {
		h.cancel()
	}
	return h
}

func convertPictureToJpg(filePath string) error {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	img, err := png.Decode(bytes.NewReader(buf))
	if err != nil {
		img, err = jpeg.Decode(bytes.NewReader(buf))
		if err != nil {
			return err
		}
	}
	newBuf := bytes.Buffer{}
	err = jpeg.Encode(&newBuf, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return err
	}
	pos := strings.LastIndex(filePath, ".")
	outputPath := filePath[:pos] + ".jpg"
	err = os.WriteFile(outputPath, newBuf.Bytes(), 0644)
	if err == nil && filePath != outputPath {
		os.Remove(filePath)
	}
	return err
}

func getUserInput(event <-chan *zero.Ctx) (string, error) {
	timer := time.NewTimer(1 * time.Minute)
	var ctx *zero.Ctx
	select {
	case ctx = <-event:
		return ctx.Event.RawMessage, nil
	case <-timer.C:
		return "", errors.New("User input timed out")
	}
}

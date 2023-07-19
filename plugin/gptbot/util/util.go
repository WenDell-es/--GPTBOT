package util

import (
	"encoding/json"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gptbot/plugin/gptbot/chat"
	"gptbot/plugin/gptbot/constants"
	"gptbot/plugin/gptbot/model"
	"gptbot/store"
	"strconv"
)

const ()

func GetChatId(ctx *zero.Ctx) int64 {
	id := ctx.Event.GroupID
	if id == 0 {
		// 用户id用负数，避免群号和qq号相同的情况出现
		id = -ctx.Event.UserID
	}
	return id
}

func StoreChat(chat *chat.Chat, id int64) error {
	buf, _ := json.Marshal(struct {
		Prompt      model.Message
		Probability int
		Model       string
	}{
		Prompt:      *chat.GetPrompt(),
		Probability: chat.GetGroupProbability(),
		Model:       chat.GetModel(),
	})
	return store.GetStoreClient().UploadObjectByBytes(buf, constants.StorePrefix+"/"+strconv.FormatInt(id, 10))
}

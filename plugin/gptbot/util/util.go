package util

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

func GetChatId(ctx *zero.Ctx) int64 {
	id := ctx.Event.GroupID
	if id == 0 {
		// 用户id用负数，避免群号和qq号相同的情况出现
		id = -ctx.Event.UserID
	}
	return id
}

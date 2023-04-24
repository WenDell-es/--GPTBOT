package util

import (
	"fmt"
	"strings"
)

func GenerateAtCQCode(userId int64) string {
	return fmt.Sprint("[CQ:at,qq=", userId, "]")
}

func CutPrefixAndTrimSpace(message string, cut string) string {
	message = strings.TrimSpace(message)
	message, _ = strings.CutPrefix(message, cut)
	message = strings.TrimSpace(message)
	return message
}

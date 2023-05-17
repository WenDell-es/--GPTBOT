package util

import (
	"fmt"
	"regexp"
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

func RemoveAllCQCode(str string) string {
	reg := regexp.MustCompile(`\[CQ:[^\]]*\]`)
	return reg.ReplaceAllString(str, "")
}

func IsStringAboutMe(str string, selfId int64) bool {
	nickName := []string{"猫娘", GenerateAtCQCode(selfId)}
	for _, name := range nickName {
		if strings.Contains(str, name) {
			return true
		}
	}
	return false
}

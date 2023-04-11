package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"gptbot/Constants"
	"gptbot/log"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type globalSetting struct {
	apiKey string
}
type userSetting struct {
	system  string
	netWork bool
	id      int
}

var global globalSetting
var users sync.Map

var logger *logrus.Logger

func myTrim(s string, cut string) string {
	for i := range s {
		if len(s[i:]) < len(cut) {
			return s
		}

		if s[i:i+len(cut)] == cut {
			add := ""
			if i+len(cut) < len(s) {
				add = s[i+len(cut):]
			}
			return s[:i] + add
		}
	}
	return s
}

func myTrimWithSpace(s string, cut string) string {
	if strings.Contains(s, cut+" ") {
		return myTrim(s, cut+" ")
	}

	return myTrim(s, cut)
}

func queryGpt(queryBody []byte) string {
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	request, err := http.NewRequest(Constants.PostMethod, Constants.GptApiUrl, bytes.NewReader(queryBody))
	if err != nil {
		logger.Errorln(err)
		return err.Error()
	}
	request.Header.Add("origin", Constants.OriginHost)
	request.Header.Add("Content-Type", Constants.DefaultContentType)

	var ans []byte
	// Default retry count is 10, You can add attempts option to change it
	err = retry.Do(func() error {
		r, err := client.Do(request)
		if err != nil {
			logger.Errorln(err, string(queryBody))
			return err
		}
		defer r.Body.Close()
		ans, err = io.ReadAll(r.Body)
		if err != nil {
			logger.Errorln(err, string(queryBody))
			return err
		}
		for _, unExpectedResp := range Constants.UnExpectedResp {
			if strings.Contains(string(ans), unExpectedResp) {
				return errors.New(unExpectedResp)
			}
		}
		return nil
	},
	)
	if err != nil {
		logger.Errorln(err)
		return err.Error()
	}
	return string(ans)
}

func send2gpt3method1(s string, id int) string {
	type req struct {
		Prompt         string `json:"prompt"`
		UserId         string `json:"userId"`
		Network        bool   `json:"network"`
		Apikey         string `json:"apikey"`
		System         string `json:"system"`
		WithoutContext bool   `json:"withoutContext"`
	}

	w := req{
		Prompt:         s,
		UserId:         "#/chat/",
		Network:        false,
		Apikey:         global.apiKey,
		System:         "",
		WithoutContext: false,
	}
	v, ok := users.Load(id)
	temp := userSetting{id: int(time.Now().UnixMilli())}
	if ok {
		temp = v.(userSetting)
	}
	users.Store(id, temp)

	w.System = temp.system
	w.Network = temp.netWork
	w.UserId += fmt.Sprint(temp.id)

	j, err := json.Marshal(w)
	if err != nil {
		logger.Errorln(err, w)
		return err.Error()
	}
	return queryGpt(j)
}

func receive(w http.ResponseWriter, r *http.Request) {
	type fromQQ struct {
		Message_type string `json:"message_type"`
		Self_id      int    `json:"self_id"`
		User_id      int    `json:"user_id"`
		Group_id     int    `json:"group_id"`
		Message      string `json:"message"`
	}
	type toQQ struct {
		Message_type string `json:"message_type"`
		User_id      int    `json:"user_id"`
		Group_id     int    `json:"group_id"`
		Message      string `json:"message"`
	}

	ans, _ := io.ReadAll(r.Body)
	r.Body.Close()

	var data fromQQ
	json.Unmarshal(ans, &data)

	go func() {
		messageType := data.Message_type
		chatId := data.User_id
		fromId := data.User_id
		myId := data.Self_id
		message := data.Message

		s := ""
		if messageType == "group" {
			if !strings.Contains(message, "[CQ:at,qq="+fmt.Sprint(myId)+"]") {
				return
			}
			message = myTrimWithSpace(message, "[CQ:at,qq="+fmt.Sprint(myId)+"]")

			s += "[CQ:at,qq=" + fmt.Sprint(fromId) + "] "

			chatId = data.Group_id
		} else if messageType == "private" {
			s = "您好，因需要技术升级，我们决定暂时关闭私聊功能"
			logger.Infoln(
				map[string]string{
					"fromId":   strconv.Itoa(fromId),
					"chatId":   strconv.Itoa(chatId),
					"question": message,
					"answer":   s,
				},
			)
			re, _ := json.Marshal(toQQ{Message_type: messageType, User_id: chatId, Group_id: chatId, Message: s})
			_, err := http.Post(`http://127.0.0.1:5700/send_msg`, "application/json", bytes.NewReader(re))
			if err != nil {
				logger.Errorln(err, w)
			}
			return
		}

		if strings.Contains(message, "-help") {
			s += `帮助：
-help 显示帮助
			
设置当前会话场景：
-set 设置提示词
-show 显示提示词
-net 联网/断网
-refresh 重置当前情景
（有的时候提示词会失效，请refresh重置，
refresh后不需要重新设定提示词）
			
设置全局：
-setkey 设置apikey
-showkey 显示apikey前后4位
-showtoken 显示余额
-showkey 显示apikey`

		} else if strings.Contains(message, "-set") {
			v, ok := users.Load(chatId)
			temp := userSetting{}
			if ok {
				temp = v.(userSetting)
			}
			temp.system = myTrimWithSpace(message, "-set")
			temp.id = int(time.Now().UnixMilli())
			users.Store(chatId, temp)

			s += "set提示词成功！重置情景成功"
		} else if strings.Contains(message, "-show") {
			v, ok := users.Load(chatId)
			if ok {
				s += v.(userSetting).system
			}
		} else if strings.Contains(message, "-net") {
			v, ok := users.Load(chatId)

			temp := userSetting{}
			if ok {
				temp = v.(userSetting)
			}

			temp.netWork = !temp.netWork
			if temp.netWork {
				s += "联网功能已打开"
			} else {
				s += "联网功能已关闭"
			}
			users.Store(chatId, temp)
		} else if strings.Contains(message, "-refresh") {
			v, ok := users.Load(chatId)

			temp := userSetting{}
			if ok {
				temp = v.(userSetting)
			}

			temp.id = int(time.Now().UnixMilli())
			users.Store(chatId, temp)

			s += "重置情景成功（-set的提示词不会被重置）"
		} else {
			s += send2gpt3method1(message, chatId)
		}

		logger.Infoln(
			map[string]string{
				"fromId":   strconv.Itoa(fromId),
				"chatId":   strconv.Itoa(chatId),
				"question": message,
				"answer":   s,
			},
		)
		re, _ := json.Marshal(toQQ{Message_type: messageType, User_id: chatId, Group_id: chatId, Message: s})
		_, err := http.Post(`http://127.0.0.1:5700/send_msg`, "application/json", bytes.NewReader(re))
		if err != nil {
			logger.Errorln(err, w)
		}
	}()
}

func main() {
	logger = log.InitLog()
	logger.Infoln("GPT Bot Start")

	global.apiKey = "ea7952ae-6262-4daf-be97-df4b4c1afe74"

	http.HandleFunc("/", receive)
	http.ListenAndServe(":5701", nil)
}

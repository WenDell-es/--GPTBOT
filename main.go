package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gptbot/log"
	"io"
	"net/http"
	"strings"
	"sync"
)

var character sync.Map
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

func send2gpt3method1(s string, id int) string {
	type req struct {
		Network bool   `json:"network"`
		Prompt  string `json:"prompt"`
		UserId  string `json:"userId"`
		Apikey  string `json:"apikey"`
	}

	w := req{false, s, "#/chat/" + fmt.Sprint(id), "8cbb290c-2c7f-44ef-9d14-df2110319da8"}
	j, err := json.Marshal(w)
	if err != nil {
		logger.Errorln(err, w)
		return err.Error()
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	for {
		r, err := client.Post("https://cbjtestapi.binjie.site:7777/api/generateStream", "application/json", bytes.NewReader(j))
		if err != nil {
			logger.Errorln(err, string(j))
			return err.Error()
		}
		ans, _ := io.ReadAll(r.Body)
		r.Body.Close()

		if !strings.Contains(string(ans), `https://chat1.yqcloud.top`) {
			return string(ans)
		}
	}
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
		if data.Message_type == "group" {
			if strings.Contains(data.Message, "[CQ:at,qq="+fmt.Sprint(data.Self_id)+"] ") {
				data.Message = myTrim(data.Message, "[CQ:at,qq="+fmt.Sprint(data.Self_id)+"] ")
			} else if strings.Contains(data.Message, "[CQ:at,qq="+fmt.Sprint(data.Self_id)+"]") {
				data.Message = myTrim(data.Message, "[CQ:at,qq="+fmt.Sprint(data.Self_id)+"]")
			} else {
				return
			}

			s := "[CQ:at,qq=" + fmt.Sprint(data.User_id) + "] "
			if strings.Contains(data.Message, "-set ") {
				data.Message = myTrim(data.Message, "-set ")
				character.Store(data.Group_id, data.Message)
				return
			} else if strings.Contains(data.Message, "-show") {
				v, ok := character.Load(data.Group_id)
				if ok {
					s += v.(string)
				}
			} else {
				v, ok := character.Load(data.Group_id)
				if ok {
					data.Message += v.(string)
				}
				s += send2gpt3method1(data.Message, data.Group_id)
			}

			logger.Infoln("qu:"+data.Message, data.Group_id, data.User_id)
			logger.Infoln("re:" + s)

			re, _ := json.Marshal(toQQ{Message_type: "group", Group_id: data.Group_id, Message: s})
			_, err := http.Post(`http://127.0.0.1:5700/send_msg`, "application/json", bytes.NewReader(re))
			if err != nil {
				logger.Errorln(err, w)
			}

		} else if data.Message_type == "private" {
			s := ""
			if strings.Contains(data.Message, "-set ") {
				data.Message = myTrim(data.Message, "-set ")
				character.Store(data.User_id, data.Message)
				return
			} else if strings.Contains(data.Message, "-show") {
				v, ok := character.Load(data.User_id)
				if ok {
					s += v.(string)
				}
			} else {
				v, ok := character.Load(data.User_id)
				if ok {
					data.Message += v.(string)
				}
				s += send2gpt3method1(data.Message, data.User_id)
			}

			logger.Infoln("qu:"+data.Message, data.User_id)
			logger.Infoln("re:" + s)

			re, _ := json.Marshal(toQQ{Message_type: "private", User_id: data.User_id, Message: s})
			_, err := http.Post(`http://127.0.0.1:5700/send_msg`, "application/json", bytes.NewReader(re))
			if err != nil {
				logger.Errorln(err, w)
			}
		}
	}()
}

func main() {
	logger = log.InitLog()
	logger.Infoln("GPT Bot Start")
	http.HandleFunc("/", receive)
	http.ListenAndServe(":5701", nil)
}

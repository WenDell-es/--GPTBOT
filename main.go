package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
)

type req struct {
	Network bool   `json:"network"`
	Prompt  string `json:"prompt"`
	UserId  string `json:"userId"`
}

type send struct {
	Nickname string `json:"nickname"`
}

type strData struct {
	Message_type string `json:"message_type"`
	Self_id      int    `json:"self_id"`
	User_id      int    `json:"user_id"`
	Group_id     int    `json:"group_id"`

	Sender send `json:"sender"`

	Message string `json:"message"`
}

type rsp struct {
	Message_type string `json:"message_type"`
	User_id      int    `json:"user_id"`
	Group_id     int    `json:"group_id"`
	Message      string `json:"message"`
}

func send2gpt(s string, id int) string {
	w := req{false, s, "#/chat/" + fmt.Sprint(id)}
	j, _ := json.Marshal(w)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	r, err := client.Post("https://cbjtestapi.binjie.site:7777/api/generateStream", "application/json", bytes.NewReader(j))

	if err != nil {
		logrus.Error(err)
		return err.Error()
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.Error(err)
		return err.Error()
	}
	return string(body)
}

func receive(w http.ResponseWriter, r *http.Request) {

	ans := make([]byte, 0)
	for {
		temp := make([]byte, 256)
		n, _ := r.Body.Read(temp)
		if n == 0 {
			break
		}

		ans = append(ans, temp[:n]...)
	}

	var data strData
	json.Unmarshal(ans, &data)

	//fmt.Println(data)

	go func() {
		for {
			if data.Message_type == "group" {
				if !strings.Contains(data.Message, "[CQ:at,qq="+fmt.Sprint(data.Self_id)+"]") {
					return
				}

				fmt.Println("qu:"+strings.Trim(strings.Trim(data.Message, "[CQ:at,qq="+fmt.Sprint(data.Self_id)+"] "), "[CQ:at,qq="+fmt.Sprint(data.Self_id)+"]"), data.Group_id)
				s := send2gpt(
					strings.Trim(strings.Trim(data.Message, "[CQ:at,qq="+fmt.Sprint(data.Self_id)+"] "), "[CQ:at,qq="+fmt.Sprint(data.Self_id)+"]"),
					data.Group_id,
				)
				fmt.Println("re:" + s)

				if !strings.Contains(s, `https://chat1.yqcloud.top`) {
					re, _ := json.Marshal(rsp{Message_type: "group", Group_id: data.Group_id, Message: s})
					http.Post(`http://127.0.0.1:5700/send_msg`, "application/json", bytes.NewReader(re))
					return
				}
			} else if data.Message_type == "private" {
				fmt.Println("qu:"+data.Message, data.User_id)
				s := send2gpt(data.Message, data.User_id)
				fmt.Println("re:" + s)

				if !strings.Contains(s, `https://chat1.yqcloud.top`) {
					re, _ := json.Marshal(rsp{Message_type: "private", User_id: data.User_id, Message: s})
					http.Post(`http://127.0.0.1:5700/send_msg`, "application/json", bytes.NewReader(re))
					return
				}
			}
		}
	}()
}

func main() {
	http.HandleFunc("/", receive)
	http.ListenAndServe(":5701", nil)
}

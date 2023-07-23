package madokapicture

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"gptbot/store"
	"net/http"
	"time"
)

type PicBuffer struct {
	Ticker     *time.Ticker
	BufferChan chan string
}

var buffer = NewPicBuffer()

func NewPicBuffer() *PicBuffer {
	pb := &PicBuffer{
		Ticker:     time.NewTicker(time.Hour * 6),
		BufferChan: make(chan string, 30),
	}
	return pb
}

func BufferInit() {
	go bufferWriter()
	go func() {
		for _ = range buffer.Ticker.C {
			Reload()
		}
	}()
}

func bufferWriter() {
	objs, err := store.GetStoreClient().FetchAllFileInfo(Storage)
	if err != nil {
		logrus.Errorln("获取对象信息错误", err)
		return
	}
	i := 0
	for {
		if i >= 100 {
			objs, err = store.GetStoreClient().FetchAllFileInfo(Storage)
			if err != nil {
				logrus.Errorln("获取对象信息错误", err)
				return
			}
			i = 0
		}
		obj := objs[getRandomNum(len(objs))]
		buffer.BufferChan <- obj.Key
		addToQQImageBuffer(store.GetStoreClient().GetObjectUrl(obj.Key))
		i++
	}
}

func Reload() {
L:
	for {
		select {
		case _, ok := <-buffer.BufferChan:
			if !ok {
				break L
			}
		default:
			break L
		}
	}
}

func (b *PicBuffer) GetUrls(n int) []string {
	urls := make([]string, n)
	for i := 0; i < n; i++ {
		key := <-b.BufferChan
		urls[i] = store.GetStoreClient().GetObjectUrl(key)
	}
	return urls
}

func addToQQImageBuffer(url string) {
	body := make(map[string]string)
	body["message_type"] = "private"
	body["user_id"] = "3550182574"
	body["message"] = "[CQ:image,file=" + url + "]"
	bytesData, _ := json.Marshal(body)
	http.Post("http://127.0.0.1:5700/send_msg", "application/json;charset=utf-8", bytes.NewBuffer(bytesData))
}

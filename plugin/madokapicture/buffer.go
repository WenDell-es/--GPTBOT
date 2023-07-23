package madokapicture

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"gptbot/store"
	"net/http"
	"sync"
	"time"
)

type PicBuffer struct {
	Urls   []string
	Mutex  sync.Mutex
	Ticker *time.Ticker
	MaxLen int
}

var buffer = NewPicBuffer()

func NewPicBuffer() *PicBuffer {
	pb := &PicBuffer{
		Urls:   []string{},
		Mutex:  sync.Mutex{},
		Ticker: time.NewTicker(time.Hour * 6),
		MaxLen: 50,
	}
	return pb
}

func BufferInit() {
	Reload(buffer)
	go func() {
		for _ = range buffer.Ticker.C {
			Reload(buffer)
		}
	}()
}

func (b *PicBuffer) addToBuffer(n int) {
	objs, err := store.GetStoreClient().FetchAllFileInfo(Storage)
	if err != nil {
		logrus.Errorln("获取对象信息错误", err)
		return
	}
	newUrls := []string{}
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		obj := objs[getRandomNum(len(objs))]
		ourl := store.GetStoreClient().GetObjectUrl(obj.Key)
		go func() {
			wg.Add(1)
			addToQQImageBuffer(ourl)
			wg.Done()
		}()
		newUrls = append(newUrls, ourl)
	}
	wg.Wait()
	buffer.Mutex.Lock()
	defer buffer.Mutex.Unlock()
	buffer.Urls = append(buffer.Urls, newUrls...)
}
func (b *PicBuffer) deleteFromBuffer(n int) {
	if n > len(b.Urls) {
		return
	}
	b.Urls = b.Urls[n:]
}

func (b *PicBuffer) GetUrls(n int) []string {
	if len(b.Urls) < n {
		b.addToBuffer(n)
	}
	buffer.Mutex.Lock()
	urls := b.Urls[0:n]
	b.deleteFromBuffer(n)
	buffer.Mutex.Unlock()
	go b.addToBuffer(n)
	return urls
}

func Reload(b *PicBuffer) {
	n := len(b.Urls)
	b.addToBuffer(b.MaxLen)
	buffer.Mutex.Lock()
	b.deleteFromBuffer(n)
	buffer.Mutex.Unlock()
}

func addToQQImageBuffer(url string) {
	body := make(map[string]string)
	body["message_type"] = "group"
	body["group_id"] = "-1"
	body["message"] = "[CQ:image,file=" + url + "]"
	bytesData, _ := json.Marshal(body)
	http.Post("http://127.0.0.1:5700/send_msg", "application/json;charset=utf-8", bytes.NewBuffer(bytesData))
}

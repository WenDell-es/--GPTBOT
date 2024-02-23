package madokapicture

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gptbot/store"
	"io/fs"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"time"
)

// PicturePool 图片缓冲池
type PicturePool struct {
	ticker            *time.Ticker     // 定时器，用于到期更新缓存
	bufferChan        chan string      // 缓冲池管道，存储已被缓存的图片URL
	frequencyTable    map[string]int64 // 图片-获取频率表
	frequencyPath     string           // 图片-获取频率表的持久化存储路径
	blockingQueueSize int
	senderURL         string
}

// NewPicBuffer 构造函数
// @Param path 缓存保存路径
func NewPicBuffer(
	path string,
	cleanDuration time.Duration,
	bufferSize int,
	blockingQueueSize int,
	senderURL string,
) *PicturePool {
	println("aaa")
	pb := &PicturePool{
		ticker:            time.NewTicker(cleanDuration),
		bufferChan:        make(chan string, bufferSize),
		frequencyTable:    newFrequency(path),
		frequencyPath:     path,
		blockingQueueSize: blockingQueueSize,
		senderURL:         senderURL,
	}
	return pb
}

// Start 启动缓冲池
func (p *PicturePool) Start() {
	// 协程，加载图片并缓冲
	go p.bufferWriter()
	// 协程，定期更新缓冲池内容
	go func() {
		for _ = range p.ticker.C {
			p.reload()
		}
	}()
}

func (p *PicturePool) GetUrls(n int) []string {
	urls := make([]string, n)
	for i := 0; i < n; i++ {
		key := <-p.bufferChan
		p.frequencyTable[key]++
		urls[i] = store.GetStoreClient().GetObjectUrl(key)
	}
	if err := p.saveFrequencyTable(); err != nil {
		logrus.Errorln(errors.Wrap(err, "保存图片频率表错误"))
	}
	return urls
}

func (p *PicturePool) GetBufferCount() int {
	return len(p.bufferChan)
}

// 清除当前缓冲池
func (p *PicturePool) reload() {
L:
	for {
		select {
		case _, ok := <-p.bufferChan:
			if !ok {
				break L
			}
		default:
			break L
		}
	}
}

// 向缓冲池中缓存图片
func (p *PicturePool) bufferWriter() {
	for {
		keys, err := p.fetchLowFrequencyObjects()
		if err != nil {
			logrus.Errorln("获取对象信息错误", err)
			return
		}
		perm := rand.Perm(len(keys))
		for i := 0; i < len(perm); i++ {
			obj := keys[i]
			p.bufferChan <- obj.Key
			p.addToQQImageBuffer(store.GetStoreClient().GetObjectUrl(obj.Key))
		}

	}
}

type keyFrequency struct {
	Key       string
	Frequency int64
}

// 获取低频图片
func (p *PicturePool) fetchLowFrequencyObjects() ([]*keyFrequency, error) {
	objs, err := store.GetStoreClient().FetchAllFileInfo(Storage)
	if err != nil {
		logrus.Errorln("获取对象信息错误", err)
		return nil, err
	}
	sortArray := make([]*keyFrequency, len(objs))
	for i, obj := range objs {
		var fre int64
		if val, ok := p.frequencyTable[obj.Key]; ok {
			fre = val
		} else {
			p.frequencyTable[obj.Key] = 0
		}
		sortArray[i] = &keyFrequency{
			Key:       obj.Key,
			Frequency: fre,
		}
	}
	sort.Slice(sortArray, func(i, j int) bool {
		return sortArray[i].Frequency < sortArray[j].Frequency
	})
	if len(sortArray) < p.blockingQueueSize {
		return sortArray, nil
	}
	return sortArray[:p.blockingQueueSize], nil
}

// 保存当前频率文件
func (p *PicturePool) saveFrequencyTable() error {
	buf, err := json.MarshalIndent(p.frequencyTable, "", "  ")
	if err != nil {
		return errors.Wrap(err, "json marshall error")
	}
	return os.WriteFile(p.frequencyPath, buf, 0644)
}

// 初始化图片频率表
func newFrequency(path string) map[string]int64 {
	ret := make(map[string]int64)
	buf, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		logrus.Fatalln("IO读取异常", err)
	}
	if buf == nil {
		return ret
	}
	err = json.Unmarshal(buf, &ret)
	if err != nil {
		logrus.Fatalln("marshal index fill error", err)
	}
	return ret
}

func (p *PicturePool) addToQQImageBuffer(url string) {
	body := make(map[string]string)
	body["message_type"] = "private"
	body["user_id"] = SelfId
	body["message"] = "[CQ:image,file=" + url + "]"
	bytesData, _ := json.Marshal(body)
	http.Post(p.senderURL, "application/json;charset=utf-8", bytes.NewBuffer(bytesData))
}

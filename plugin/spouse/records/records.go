package records

import (
	"gptbot/plugin/spouse/model"
	"sync"
	"time"
)

var (
	instance *SpouseRecorder
	once     sync.Once
)

type SpouseRecorder struct {
	records sync.Map
}

type typeGroupRecord map[model.Type]map[int64]record

type record struct {
	Date time.Time
	Card *model.Card
}

func GetSpouseRecorder() *SpouseRecorder {
	once.Do(func() {
		instance = &SpouseRecorder{records: sync.Map{}}
	})
	return instance
}

func (r *SpouseRecorder) HasSpouseToday(userId, groupId int64, spouseType model.Type) bool {
	if rdMap, ok := r.records.Load(userId); ok &&
		rdMap.(typeGroupRecord)[spouseType][groupId].Date.Format("20060102") == time.Now().Format("20060102") {
		return true
	}
	return false
}

func (r *SpouseRecorder) GetSpouseToday(userId, groupId int64, spouseType model.Type) *model.Card {
	if rdMap, ok := r.records.Load(userId); ok &&
		rdMap.(typeGroupRecord)[spouseType][groupId].Date.Format("20060102") == time.Now().Format("20060102") {
		return rdMap.(typeGroupRecord)[spouseType][groupId].Card
	}
	return nil
}

func (r *SpouseRecorder) AddSpouseToday(userId, groupId int64, spouseType model.Type, card *model.Card) {
	rdMap := make(typeGroupRecord)
	if rd, ok := r.records.Load(userId); ok {
		rdMap = rd.(typeGroupRecord)
	}
	stMap := make(map[int64]record)
	if st, ok := rdMap[spouseType]; ok {
		stMap = st
	}
	stMap[groupId] = record{
		Date: time.Now(),
		Card: card,
	}
	rdMap[spouseType] = stMap
	r.records.Store(userId, rdMap)
}

package records

import (
	"gptbot/plugin/spouse/model"
	"sync"
	"time"
)

var UserRecords = sync.Map{}

type Record struct {
	Date time.Time
	Card model.Card
}

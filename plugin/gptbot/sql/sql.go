package sql

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type chat struct {
	gorm.Model
	Gid         int64 `gorm:"primaryKey"`
	GptModel    string
	Probability int
	Prompt      string
}

type Storage struct {
	db *gorm.DB
}

func NewStorage(dsName string) *Storage {
	return &Storage{
		db: newDb(dsName),
	}
}

func newDb(dsName string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsName), &gorm.Config{})
	if err != nil {
		logrus.Fatalln(err, "failed to connect database")
	}
	err = db.AutoMigrate(&chat{})
	if err != nil {
		logrus.Fatalln(err, "failed to connect database")
	}
	return db
}

func (s *Storage) CreateChat(gid int64, model string, pro int, prompt string) {
	s.db.Create(&chat{
		Gid:         gid,
		GptModel:    model,
		Probability: pro,
		Prompt:      prompt,
	})
}
func (s *Storage) FindAllChats() []chat {
	var chats []chat
	s.db.Find(&chats)
	return chats
}

func (s *Storage) UpdateModel(gid int64, model string) {
	var c chat
	s.db.First(&c, gid)
	s.db.Model(&c).Update("GptModel", model)
}

func (s *Storage) UpdateProbability(gid int64, probability int) {
	var c chat
	s.db.First(&c, gid)
	s.db.Model(&c).Update("Probability", probability)
}

func (s *Storage) UpdatePrompt(gid int64, prompt string) {
	var c chat
	s.db.First(&c, gid)
	s.db.Model(&c).Update("Prompt", prompt)
}

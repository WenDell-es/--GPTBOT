package log

import "github.com/sirupsen/logrus"

type EventFieldHook struct {
}

func (hook *EventFieldHook) Fire(entry *logrus.Entry) error {
	return nil
}

func (hook *EventFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func NewEventFieldHook() logrus.Hook {
	return &EventFieldHook{}
}

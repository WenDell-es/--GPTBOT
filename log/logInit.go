package log

import (
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	logPath         = "./logs/log"
	maxLogAge       = time.Duration(30*24) * time.Hour
	logRotationTime = time.Duration(24) * time.Hour
)

func LogInit() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     true,
	})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetReportCaller(true)
	writer, err := rotateLogs.New(
		logPath+".%Y%m%d",
		rotateLogs.WithLinkName(logPath),
		rotateLogs.WithMaxAge(maxLogAge),
		rotateLogs.WithRotationTime(logRotationTime),
	)
	if err != nil {
		panic("Log init failed" + err.Error())
	}
	logrus.SetOutput(writer)
	logrus.AddHook(NewEventFieldHook())
}

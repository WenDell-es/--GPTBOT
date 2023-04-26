package log

import (
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	logPath         = "./gpt_bot_logs/log"
	maxLogAge       = time.Duration(30*24) * time.Hour
	logRotationTime = time.Duration(24) * time.Hour
)

func InitLog() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     true,
	})
	logger.SetLevel(logrus.InfoLevel)
	logger.SetReportCaller(true)
	writer, err := rotateLogs.New(
		logPath+".%Y%m%d",
		rotateLogs.WithLinkName(logPath),
		rotateLogs.WithMaxAge(maxLogAge),
		rotateLogs.WithRotationTime(logRotationTime),
	)
	if err != nil {
		panic("Log init failed" + err.Error())
	}
	logger.SetOutput(writer)
	return logger
}

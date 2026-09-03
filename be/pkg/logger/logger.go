package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger(logLevel, appEnv string) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	// Configure formatter based on environment
	if strings.ToLower(appEnv) == "production" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	}

	// Parse log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	Log = log
	return log
}

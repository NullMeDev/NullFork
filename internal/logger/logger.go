package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// InitLogger initializes the logger based on the provided configuration
func InitLogger(logLevel, logFormat string, includeCaller bool) *logrus.Logger {
	log := logrus.New()
	log.Out = os.Stdout

	switch logLevel {
	case "DEBUG":
		log.SetLevel(logrus.DebugLevel)
	case "INFO":
		log.SetLevel(logrus.InfoLevel)
	case "WARN":
		log.SetLevel(logrus.WarnLevel)
	case "ERROR":
		log.SetLevel(logrus.ErrorLevel)
	case "FATAL":
		log.SetLevel(logrus.FatalLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	if logFormat == "JSON" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else if logFormat == "TEXT" {
		formatter := &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		}
		log.SetFormatter(formatter)
	}

	if includeCaller {
		log.SetReportCaller(true)
	}

	return log
}


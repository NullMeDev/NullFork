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
	case "TRACE":
		log.SetLevel(logrus.TraceLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	if logFormat == "JSON" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			DisableTimestamp: false,
			DisableHTMLEscape: false,
			DataKey: "fields",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
				logrus.FieldKeyFile:  "file",
			},
		})
	} else if logFormat == "TEXT" {
		formatter := &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			ForceColors:     true,
			DisableColors:   false,
			DisableQuote:    false,
			QuoteEmptyFields: true,
		}
		log.SetFormatter(formatter)
	}

	if includeCaller {
		log.SetReportCaller(true)
	}

	// Add context fields for better tracing
	log.WithFields(logrus.Fields{
		"service": "enhanced-gateway-scraper",
		"version": "1.0.0",
		"pid": os.Getpid(),
	})

	return log
}


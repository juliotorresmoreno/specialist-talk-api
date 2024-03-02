package logger

import (
	"os"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

func SetupLogrus() {
	logrus.SetOutput(colorable.NewColorableStdout())

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	logrus.SetLevel(logrus.InfoLevel)

	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func SetupLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetOutput(colorable.NewColorableStdout())

	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	logger.SetLevel(logrus.InfoLevel)

	if os.Getenv("LOGGER") == "all" {
		logger.SetReportCaller(true)
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	return logger
}

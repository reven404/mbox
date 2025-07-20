package main

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func setupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	
	logger.SetOutput(os.Stdout)
	
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})

	level, err := logrus.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		logger.SetLevel(logrus.InfoLevel)
		logger.Warnf("Invalid log level '%s', defaulting to 'info'", logLevel)
	} else {
		logger.SetLevel(level)
	}

	return logger
}
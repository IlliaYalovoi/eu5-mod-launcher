package logging

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger = newLogger()

func newLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
		ForceColors:     true,
		PadLevelText:    true,
	})
	l.SetLevel(parseLevel(os.Getenv("EU5_LOG_LEVEL")))
	return l
}

func parseLevel(raw string) logrus.Level {
	if strings.TrimSpace(raw) == "" {
		return logrus.InfoLevel
	}
	level, err := logrus.ParseLevel(strings.ToLower(strings.TrimSpace(raw)))
	if err != nil {
		return logrus.InfoLevel
	}
	return level
}

func Debugf(format string, args ...any) {
	logger.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	logger.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	logger.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	logger.Errorf(format, args...)
}

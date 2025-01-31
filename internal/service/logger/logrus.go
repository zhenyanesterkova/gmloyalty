package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	LogrusLog *logrus.Logger
}

func NewLogrusLogger() LogrusLogger {
	return LogrusLogger{
		LogrusLog: logrus.New(),
	}
}

func (llog LogrusLogger) SetLevelForLog(level string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("logruslogger.go func SetLevelForLog(): error parse log level - %w", err)
	}

	llog.LogrusLog.SetLevel(lvl)
	return nil
}

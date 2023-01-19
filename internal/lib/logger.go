package lib

import (
	"github.com/nekomeowww/perobot/pkg/logger"
	"github.com/sirupsen/logrus"
)

func NewLogger() func() *logger.Logger {
	return func() *logger.Logger {
		return logger.NewLogger(logrus.InfoLevel, "perobot", "", make([]logrus.Hook, 0))
	}
}

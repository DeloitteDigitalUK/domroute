package config

import (
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

func InitLogger() {
	level, _ := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if level == 0 {
		level = logrus.DebugLevel
	}
	logrus.SetLevel(level)
}

func GetCheckInterval() time.Duration {
	checkInterval := os.Getenv("CHECK_INTERVAL")

	if checkInterval == "" {
		return 60 * time.Second
	} else {
		interval, err := strconv.Atoi(checkInterval)
		if err != nil {
			panic(err)
		}
		return time.Duration(interval) * time.Second
	}
}

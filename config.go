package main

import (
	"os"
	"strconv"
	"time"
)

func getCheckInterval() time.Duration {
	checkInterval := os.Getenv("CHECK_INTERVAL")

	if checkInterval == "" {
		return 30 * time.Second
	} else {
		interval, err := strconv.Atoi(checkInterval)
		if err != nil {
			panic(err)
		}
		return time.Duration(interval) * time.Second
	}
}

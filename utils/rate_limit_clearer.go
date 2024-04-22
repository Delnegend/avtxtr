package utils

import (
	"log/slog"
	"os"
	"time"
)

func RateLimitClearer(ipList *map[string]int) {
	clearListEveryTimeUnit := 1 * time.Hour
	if envValue := os.Getenv("CLEAR_LIST_EVERY_TIME_UNIT"); envValue != "" {
		if parsedValue, err := time.ParseDuration(envValue); err != nil {
			slog.Warn("CLEAR_LIST_EVERY_TIME_UNIT is not a valid duration, using default value", "value", clearListEveryTimeUnit)
		} else {
			clearListEveryTimeUnit = parsedValue
			slog.Info("CLEAR_LIST_EVERY_TIME_UNIT is set", "value", clearListEveryTimeUnit)
		}
	} else {
		slog.Warn("CLEAR_LIST_EVERY_TIME_UNIT is not set, using default value", "value", clearListEveryTimeUnit)
	}

	for {
		*ipList = make(map[string]int)

		time.Sleep(clearListEveryTimeUnit)
	}
}

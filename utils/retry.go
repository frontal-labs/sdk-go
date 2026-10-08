package utils

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RetryAfter parses an HTTP Retry-After value. It returns false when the value is invalid.
func RetryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		if seconds < 0 || seconds > int64((1<<63-1)/int64(time.Second)) {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}
	date, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	delay := date.Sub(now)
	if delay < 0 {
		delay = 0
	}
	return delay, true
}

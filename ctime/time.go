package ctime

import (
	"time"
)

func NowUTC() time.Time {
	return time.Now().UTC()
}

func NowUTCMilli() int64 {
	return time.Now().UTC().UnixMilli()
}

func TimeZero() time.Time {
	return time.Time{}
}

func UnixDays(t time.Time) int64 {
	return t.Unix() / 86400
}

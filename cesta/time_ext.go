package cesta

import "time"

func nowUTC() time.Time {
	return time.Now().UTC()
}

func nowUTCMilli() int64 {
	return time.Now().UTC().UnixMilli()
}

func nowUTCDays() int64 {
	return time.Now().UTC().Unix() / 86400
}

func unixDays(t time.Time) int64 {
	return t.Unix() / 86400
}

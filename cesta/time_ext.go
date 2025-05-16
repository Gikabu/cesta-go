package cesta

import "time"

func nowUTC() time.Time {
	return time.Now().UTC()
}

func nowUTCMilli() int64 {
	return time.Now().UTC().UnixMilli()
}

func nowUnixDays() int64 {
	return unixDays(nowUTC())
}

func unixDays(t time.Time) int64 {
	return t.Unix() / 86400
}

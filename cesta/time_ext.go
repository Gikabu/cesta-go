package cesta

import "time"

func nowUTC() time.Time {
	return time.Now().UTC()
}

func nowUTCMilli() int64 {
	return time.Now().UTC().UnixMilli()
}

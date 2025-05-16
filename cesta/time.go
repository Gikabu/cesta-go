package cesta

import "time"

func GetCurrentBase34Date() (string, error) {
	return TryEncodeBase34(nowUnixDays())
}

func CurrentBase34Date() string {
	code, err := TryEncodeBase34(nowUnixDays())
	if err != nil {
		panic(err)
	}
	return code
}

func NewBase34Date(t time.Time) string {
	code, err := TryEncodeBase34(unixDays(t))
	if err != nil {
		panic(err)
	}
	return code
}

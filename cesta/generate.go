package cesta

import "strconv"

// TryGenerateBase34Code - Generate a random base34 code of the length specified
// It returns an error if unsuccessful
func TryGenerateBase34Code(length int) (string, error) {
	return generateRandomBase34(length)
}

// GenerateBase34Code - Generate a random base34 code of the length specified
// It panics if unsuccessful
func GenerateBase34Code(length int) string {
	code, err := generateRandomBase34(length)
	if err != nil {
		panic(err)
	}
	return code
}

// DefaultBase34Code - Generate a random base34 code of default length (8)
// If unsuccessful, it returns the last eight digits of nanoseconds elapsed since January 1, 1970, UTC
func DefaultBase34Code() string {
	code, err := generateRandomBase34(8)
	if err != nil {
		var nano = strconv.FormatInt(nowUTC().UnixNano(), 10)
		return nano[len(nano)-8:]
	}
	return code
}

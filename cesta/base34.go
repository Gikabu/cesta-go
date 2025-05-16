package cesta

// TryEncodeBase34 - encodes a 64-bit integer to a base34 string
// It returns an error if unsuccessful
func TryEncodeBase34(number int64) (string, error) {
	intBytes, err := int64ToBytes(number)
	if err != nil {
		return "", err
	}
	return encodeBase34(intBytes), nil
}

// EncodeBase34 - encodes a 64-bit integer to a base34 string
// It panics if unsuccessful
func EncodeBase34(number int64) string {
	base34, err := TryEncodeBase34(number)
	if err != nil {
		panic(err)
	}
	return base34
}

// TryDecodeBase34 - decodes a base34-encoded string to a 64-bit integer
// It returns an error if unsuccessful
func TryDecodeBase34(encoded string) (int64, error) {
	base34, err := decodeBase34(encoded)
	if err != nil {
		return 0, err
	}
	return bytesToInt64(base34)
}

// DecodeBase34 - decodes a base34-encoded string to a 64-bit integer
// It panics if unsuccessful
func DecodeBase34(encoded string) int64 {
	number, err := TryDecodeBase34(encoded)
	if err != nil {
		panic(err)
	}
	return number
}

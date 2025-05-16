package cesta

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"
)

const base34Chars = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ" // Excludes 'I' and 'O'

func encodeBase34(input []byte) string {
	num := new(big.Int).SetBytes(input)
	base := big.NewInt(34)
	zero := big.NewInt(0)
	var encoded strings.Builder

	for num.Cmp(zero) > 0 {
		mod := new(big.Int)
		num.DivMod(num, base, mod)
		encoded.WriteString(string(base34Chars[mod.Int64()]))
	}

	result := encoded.String()
	// Reverse the encoded string
	runes := []rune(result)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func decodeBase34(encoded string) ([]byte, error) {
	base := big.NewInt(34)
	num := big.NewInt(0)

	for _, char := range encoded {
		index := strings.IndexRune(base34Chars, char)
		if index == -1 {
			return nil, fmt.Errorf("invalid character '%c' in base34 string", char)
		}
		num.Mul(num, base)
		num.Add(num, big.NewInt(int64(index)))
	}

	return num.Bytes(), nil
}

func int64ToBytes(num int64) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func bytesToInt64(data []byte) (int64, error) {
	var num int64
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, &num)
	if err != nil {
		return 0, err
	}
	return num, nil
}

// generateRandomBase34 generates a random Base34 string of specified length
func generateRandomBase34(length int) (string, error) {
	var result []byte
	count := big.NewInt(int64(len(base34Chars)))

	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, count)
		if err != nil {
			return "", err
		}
		result = append(result, base34Chars[index.Int64()])
	}
	return string(result), nil
}

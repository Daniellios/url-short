package service

import (
	"crypto/rand"
	"fmt"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func ShortenURL(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than 0")
	}

	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	code := make([]byte, length)
	for i, b := range buf {
		code[i] = base62Alphabet[int(b)%len(base62Alphabet)]
	}

	return string(code), nil
}

package service

import (
	"math/rand"
)

func generationShortLink() string {
	var (
		charset string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		length  int    = 6
	)

	sliceByte := make([]byte, length)

	for i := 0; i < length; i++ {
		sliceByte[i] = charset[rand.Intn(len(charset))]
	}

	return string(sliceByte)
}

package utils

import (
	"math/rand"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must2[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnouvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}

	return string(s)
}

func Ptr[T any](val T) *T {
	return &val
}

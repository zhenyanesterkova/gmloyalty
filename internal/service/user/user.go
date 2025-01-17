package user

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	sizeSalt    = 8
	hashTime    = 1
	hashMemory  = 64 * 1024
	hashThreads = 4
	hashKeyLen  = 32
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u User) HashPassword() (string, error) {
	salt, err := createSalt(sizeSalt)
	if err != nil {
		return "", fmt.Errorf("failed generate salt for calc hash password: %w", err)
	}

	hashedPass := argon2.IDKey([]byte(u.Password), salt, hashTime, hashMemory, hashThreads, hashKeyLen)

	res := []byte{}
	res = append(res, salt...)
	res = append(res, hashedPass...)

	return hex.EncodeToString(res), nil
}

func createSalt(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return []byte{}, fmt.Errorf("failed generating random bytes: %w", err)
	}

	return b, nil
}

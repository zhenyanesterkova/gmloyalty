package user

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
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

var (
	ErrBadPass = errors.New("bad password")
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Accaunt struct {
	ID        int     `json:"-"`
	UserID    int     `json:"-"`
	Balance   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (u User) CheckPassword(hashPasswordDB string) error {
	passDB, err := hex.DecodeString(hashPasswordDB)
	if err != nil {
		return fmt.Errorf("failed decode hash password from DB to []bytes: %w", err)
	}

	salt := passDB[:sizeSalt]

	hashPassFromQuery, err := u.HashPassword(salt)
	if err != nil {
		return fmt.Errorf("failed calc hash password: %w", err)
	}
	hashPassFromQueryBytes, err := hex.DecodeString(hashPassFromQuery)
	if err != nil {
		return fmt.Errorf("failed decode hash password from query to []bytes: %w", err)
	}

	if !bytes.Equal(hashPassFromQueryBytes, passDB) {
		return ErrBadPass
	}
	return nil
}

func (u User) HashPassword(salt []byte) (string, error) {
	hashedPass := argon2.IDKey([]byte(u.Password), salt, hashTime, hashMemory, hashThreads, hashKeyLen)

	res := []byte{}
	res = append(res, salt...)
	res = append(res, hashedPass...)

	return hex.EncodeToString(res), nil
}

func CreateSalt() ([]byte, error) {
	b := make([]byte, sizeSalt)
	_, err := rand.Read(b)
	if err != nil {
		return []byte{}, fmt.Errorf("failed generating random bytes: %w", err)
	}

	return b, nil
}

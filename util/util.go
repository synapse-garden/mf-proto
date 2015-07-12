package util

import (
	"crypto/sha1"
	"time"

	uuid "github.com/streadway/simpleuuid"
	k "golang.org/x/crypto/pbkdf2"
)

type Key []byte

func PwHashKey(pwhash []byte) (Key, error) {
	salt, err := uuid.NewTime(time.Now())
	if err != nil {
		return Key{}, err
	}
	return Key(k.Key(pwhash, salt.Bytes(), 4096, 32, sha1.New)), nil
}

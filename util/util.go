package util

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
)

type Key string
type Hash string
type Salt string

func CheckHashedPw(pw string, salt Salt, hash Hash) bool {
	hashedPw := sha1.Sum([]byte(pw))
	sha := sha1.New()
	io.WriteString(sha, string(salt))
	io.WriteString(sha, hex.EncodeToString(hashedPw[:]))
	return hex.EncodeToString(sha.Sum(nil)[:]) == string(hash)
}

func HashedAndSalt(pw, saltSeed string) (Hash, Salt) {
	hashedPw := sha1.Sum([]byte(pw))
	sha := sha1.New()
	hashedSalt := sha1.Sum([]byte(saltSeed))
	io.WriteString(sha, hex.EncodeToString(hashedSalt[:]))
	io.WriteString(sha, hex.EncodeToString(hashedPw[:]))
	return Hash(hex.EncodeToString(sha.Sum(nil)[:])),
		Salt(hex.EncodeToString(hashedSalt[:]))
}

func SaltedHash(pw, saltSeed string) Key {
	hashedPw := sha1.Sum([]byte(pw))
	sha := sha1.New()
	hashedSalt := sha1.Sum([]byte(saltSeed))
	io.WriteString(sha, hex.EncodeToString(hashedSalt[:]))
	io.WriteString(sha, hex.EncodeToString(hashedPw[:]))
	return Key(hex.EncodeToString(sha.Sum(nil)[:]))
}

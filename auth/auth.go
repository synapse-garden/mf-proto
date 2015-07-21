package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/util"
)

type b []byte

const (
	SessionKeys db.Bucket = "auth-session-keys"
	Users       db.Bucket = "auth-users"
)

type User struct {
	Email string    `json:"email",omitempty"`
	Salt  util.Salt `json:"salt,omitempty"`
	Hash  util.Hash `json:"hash,omitempty"`
	Key   util.Key  `json:"key",omitempty"`
}

func Buckets() []db.Bucket {
	return []db.Bucket{
		SessionKeys,
		Users,
	}
}

func CreateUser(d db.DB, email, pwhash string) error {
	seed := time.Now().String()
	hash, salt := util.HashedAndSalt(pwhash, seed)
	return db.StoreKeyValue(d, Users, b(email), User{
		Email: email,
		Salt:  salt,
		Hash:  hash,
	})
}

func Valid(d db.DB, email string, key util.Key) error {
	userBytes, err := db.GetByKey(d, SessionKeys, []byte(email))
	if err != nil {
		return err
	}

	if len(userBytes) == 0 {
		return errors.UserNotFoundf("user %q not logged in", email)
	}

	var k util.Key
	err = json.Unmarshal(userBytes, &k)
	if err != nil {
		return err
	}

	if k == key {
		return nil
	}

	return fmt.Errorf("bad key %q for user %q", key, email)
}

func LoginUser(d db.DB, email, pwhash string) (util.Key, error) {
	// Get the hash and password for the email.
	// util.CheckHashedPw(pw, salt, hash)
	// if ok, then log in.
	userBytes, err := db.GetByKey(d, Users, []byte(email))
	if err != nil {
		return "", err
	}

	if len(userBytes) == 0 {
		return "", fmt.Errorf("no user for email %q", email)
	}

	var u User
	err = json.Unmarshal(userBytes, &u)
	if err != nil {
		return "", err
	}

	ok := util.CheckHashedPw(pwhash, u.Salt, u.Hash)
	if !ok {
		return "", fmt.Errorf("invalid password")
	}

	key := util.SaltedHash(pwhash, time.Now().String())
	err = db.StoreKeyValue(d, SessionKeys, b(email), key)
	if err != nil {
		return "", err
	}

	return key, nil
}

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
	SessionKeys db.Bucket     = "auth-session-keys"
	Users       db.Bucket     = "auth-users"
	Timeout     time.Duration = time.Duration(5) * time.Minute
)

type User struct {
	Email string    `json:"email,omitempty"`
	Salt  util.Salt `json:"salt,omitempty"`
	Hash  util.Hash `json:"hash,omitempty"`
	Key   util.Key  `json:"key,omitempty"`
}

type Login struct {
	Key     util.Key  `json:"key,omitempty"`
	Timeout time.Time `json:"timeout,omitempty"`
}

func Buckets() []db.Bucket {
	return []db.Bucket{
		SessionKeys,
		Users,
	}
}

func CreateUser(d db.DB, email, pwhash string) error {
	userBytes, err := db.GetByKey(d, Users, []byte(email))
	if err != nil {
		return err
	}

	if len(userBytes) != 0 {
		return errors.AlreadyExistsf("user for email %q", email)
	}

	seed := time.Now().String()
	hash, salt := util.HashedAndSalt(pwhash, seed)
	return db.StoreKeyValue(d, Users, b(email), User{
		Email: email,
		Salt:  salt,
		Hash:  hash,
	})
}

func DeleteUser(d db.DB, email string) error {
	userBytes, err := db.GetByKey(d, Users, []byte(email))
	if err != nil {
		return err
	}

	if len(userBytes) != 0 {
		err := db.DeleteByKey(d, Users, b(email))
		if err != nil {
			return errors.Annotatef(err, "failed to delete user %q", email)
		}
	}

	err = db.DeleteByKey(d, SessionKeys, b(email))
	if err != nil {
		return errors.Annotatef(err, "failed to log out deleted user %q", email)
	}
	return nil
}

func Valid(d db.DB, email string, key util.Key) error {
	userBytes, err := db.GetByKey(d, SessionKeys, []byte(email))
	if err != nil {
		return err
	}

	if len(userBytes) == 0 {
		return errors.UserNotFoundf("user %q not logged in", email)
	}

	var login Login
	err = json.Unmarshal(userBytes, &login)
	if err != nil {
		return err
	}

	if login.Key == key {
		t := time.Now()
		if t.Before(login.Timeout) {
			err = db.StoreKeyValue(d, SessionKeys, b(email), Login{key, t.Add(Timeout)})
			return nil
		}
		return errors.NotValidf("user %q timed out", email)
	}

	return fmt.Errorf("bad key %q for user %q", key, email)
}

func LogoutUser(d db.DB, email string, key util.Key) error {
	err := Valid(d, email, key)
	if err != nil {
		return err
	}

	err = db.DeleteByKey(d, SessionKeys, b(email))
	if err != nil {
		return err
	}

	return nil
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
	timeout := time.Now().Add(Timeout)
	err = db.StoreKeyValue(d, SessionKeys, b(email), Login{key, timeout})
	if err != nil {
		return "", err
	}

	return key, nil
}

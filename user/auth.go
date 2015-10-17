package user

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/util"
)

const (
	LoginKeys db.Bucket = "user-login-keys"
)

var timeout = struct {
	t time.Duration
	sync.RWMutex
}{t: time.Duration(5) * time.Minute}

func GetTimeout() time.Duration {
	timeout.RLock()
	defer timeout.RUnlock()
	return timeout.t
}

func SetTimeout(t time.Duration) {
	timeout.Lock()
	defer timeout.Unlock()
	timeout.t = t
}

type Login struct {
	Key     util.Key  `json:"key,omitempty"`
	Timeout time.Time `json:"timeout,omitempty"`
}

func ValidLogin(d db.DB, email string, key util.Key) error {
	login, err := GetLogin(d, email)
	if err != nil {
		return errors.NotValidf("could not get login for email %q", email)
	}

	if login.Key == key {
		t := time.Now()
		if t.Before(login.Timeout) {
			err = db.StoreKeyValue(d, LoginKeys, []byte(email), Login{key, t.Add(GetTimeout())})
			return nil
		}
		return errors.NotValidf("user %q timed out", email)
	}

	return errors.NotValidf("bad key %q for user %q", key, email)
}

func LogoutUser(d db.DB, email string, key util.Key) error {
	err := ValidLogin(d, email, key)
	if err != nil {
		return err
	}

	return ClearLogin(d, email)
}

func LoginUser(d db.DB, email, pwhash string) (util.Key, error) {
	if err := CheckUser(d, email, pwhash); err != nil {
		return "", err
	}

	key := util.SaltedHash(pwhash, time.Now().String())
	timeout := time.Now().Add(GetTimeout())

	err := db.StoreKeyValue(
		d,
		LoginKeys,
		[]byte(email),
		Login{key, timeout},
	)
	if err != nil {
		return "", err
	}

	return key, nil
}

func CheckUser(d db.DB, email, pwhash string) error {
	u, err := Get(d, email)
	if err != nil {
		return err
	}

	if ok := util.CheckHashedPw(pwhash, u.Salt, u.Hash); !ok {
		return fmt.Errorf("invalid password")
	}

	return nil
}

func GetLogin(d db.DB, email string) (*Login, error) {
	loginBytes, err := db.GetByKey(d, LoginKeys, []byte(email))
	if err != nil {
		return nil, err
	}

	if len(loginBytes) == 0 {
		return nil, errors.UserNotFoundf("user %q not logged in: ", email)
	}

	login := &Login{}
	err = json.Unmarshal(loginBytes, &login)
	if err != nil {
		return nil, err
	}
	return login, nil
}

func ClearLogin(d db.DB, email string) error {
	loginBytes, err := db.GetByKey(d, LoginKeys, []byte(email))
	if err != nil {
		return err
	}

	if len(loginBytes) == 0 {
		return errors.UserNotFoundf("user %q not logged in: ", email)
	}

	login := &Login{}
	if err = json.Unmarshal(loginBytes, &login); err != nil {
		return err
	}

	return db.DeleteByKey(d, LoginKeys, []byte(email))
}

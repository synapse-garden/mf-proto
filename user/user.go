package user

import (
	"encoding/json"
	"time"

	"github.com/juju/errors"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/object"
	"github.com/synapse-garden/mf-proto/util"
)

type b []byte

const (
	Users db.Bucket = "user-users"
)

type User struct {
	Email string    `json:"email,omitempty"`
	Salt  util.Salt `json:"salt,omitempty"`
	Hash  util.Hash `json:"hash,omitempty"`
	Key   util.Key  `json:"key,omitempty"`
}

func Buckets() []db.Bucket {
	return []db.Bucket{
		LoginKeys,
		Users,
	}
}

func Create(d db.DB, email, pwhash string) error {
	userBytes, err := db.GetByKey(d, Users, []byte(email))

	switch {
	case err != nil:
		return err
	case len(userBytes) != 0:
		return errors.AlreadyExistsf("user for email %q", email)
	}

	seed := time.Now().String()
	hash, salt := util.HashedAndSalt(pwhash, seed)
	return db.StoreKeyValue(d, Users, []byte(email), User{
		Email: email,
		Salt:  salt,
		Hash:  hash,
	})
}

func Delete(d db.DB, email string) error {
	userBytes, err := db.GetByKey(d, Users, []byte(email))
	if err != nil {
		return err
	}

	if len(userBytes) == 0 {
		return errors.Errorf("user for email %q not found", email)
	}

	if err = db.DeleteByKey(d, Users, []byte(email)); err != nil {
		return errors.Annotatef(err, "failed to delete user %q", email)
	}

	if err = object.DeleteAll(d, email); err != nil {
		return errors.Annotatef(err, "failed to clear objects for user %q", email)
	}

	if _, err := GetLogin(d, email); err == nil {
		return ClearLogin(d, email)
	}

	return nil
}

func Get(d db.DB, email string) (*User, error) {
	userBytes, err := db.GetByKey(d, Users, []byte(email))
	if err != nil {
		return nil, err
	}
	if len(userBytes) == 0 {
		return nil, errors.UserNotFoundf("%q", email)
	}

	u := &User{}
	err = json.Unmarshal(userBytes, &u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

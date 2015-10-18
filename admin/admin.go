package admin

import (
	"encoding/json"
	"time"

	"github.com/juju/errors"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/user"
	"github.com/synapse-garden/mf-proto/util"
)

const (
	admins db.Bucket = "admin-admins"
	emails db.Bucket = "admin-emails"
)

// Buckets gets the Buckets for admin.
func Buckets() []db.Bucket {
	return []db.Bucket{
		admins,
		emails,
	}
}

// Admin is a user as considered by the admin package.
type Admin user.User

// IsAdmin returns nil if there exists an Admin for the given util.Key.
func IsAdmin(d db.DB, key util.Key) error {
	if _, err := Get(d, key); err != nil {
		return err
	}

	return nil
}

// IsAdminEmail returns nil if there exists an Admin for the given email.
func IsAdminEmail(d db.DB, email string) error {
	if _, err := GetByEmail(d, email); err != nil {
		return err
	}

	return nil
}

// Get retrieves an *Admin from the database for a given key.
func Get(d db.DB, key util.Key) (*Admin, error) {
	adminJSON, err := db.GetByKey(d, admins, []byte(key))

	switch {
	case err != nil:
		return nil, err
	case len(adminJSON) == 0:
		return nil, errors.UserNotFoundf("admin for key %s:", key)
	}

	admin := new(Admin)
	err = json.Unmarshal(adminJSON, admin)
	return admin, err
}

// GetByEmail retrieves an *Admin from the database for a given email.
func GetByEmail(d db.DB, email string) (*Admin, error) {
	adminJSON, err := db.GetByKey(d, emails, []byte(email))

	switch {
	case err != nil:
		return nil, err
	case len(adminJSON) == 0:
		return nil, errors.UserNotFoundf("admin for email %s:", email)
	}

	admin := new(Admin)
	err = json.Unmarshal(adminJSON, admin)
	return admin, err
}

// Create makes a new Admin account with a given email and pwhash.
func Create(d db.DB, email, pwhash string) (util.Key, error) {
	var none util.Key
	adminJSON, err := db.GetByKey(d, emails, []byte(email))

	switch {
	case err != nil:
		return none, err
	case len(adminJSON) != 0:
		return none, errors.AlreadyExistsf("admin for email %s:", email)
	}

	hash, salt := util.HashedAndSalt(pwhash, time.Now().String())
	seed := time.Now().String()
	key := util.SaltedHash(string(hash), seed)

	adm := &Admin{
		Email: email,
		Salt:  salt,
		Hash:  hash,
		Key:   key,
	}

	if err := db.StoreKeyValue(d, admins, []byte(key), adm); err != nil {
		return none, err
	}

	return key, db.StoreKeyValue(d, emails, []byte(email), adm)
}

// Delete deletes the admin which has the given key.
func Delete(d db.DB, key util.Key) error {
	adminJSON, err := db.GetByKey(d, admins, []byte(key))

	switch {
	case err != nil:
		return err
	case len(adminJSON) == 0:
		return errors.UserNotFoundf("admin for key %s:", key)
	}

	adm := new(Admin)
	if err := json.Unmarshal(adminJSON, adm); err != nil {
		return err
	}

	if err := db.DeleteByKey(d, admins, []byte(key)); err != nil {
		return err
	}

	return db.DeleteByKey(d, emails, []byte(adm.Email))
}

// DeleteByEmail deletes the admin which has the given email.
func DeleteByEmail(d db.DB, email string) error {
	adm, err := GetByEmail(d, email)
	if err != nil {
		return err
	}

	if err := db.DeleteByKey(d, admins, []byte(adm.Key)); err != nil {
		return err
	}

	return db.DeleteByKey(d, emails, []byte(email))
}

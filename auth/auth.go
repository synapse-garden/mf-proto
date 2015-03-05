package auth

import (
	"github.com/boltdb/bolt"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/util"
)

type Auther interface {
	StoreKey(email, key []byte) error
}

func AuthUser(d db.DB, email, pwhash string) (util.Key, error) {
	key, err := util.PwHashKey([]byte(pwhash))
	if err != nil {
		return nil, err
	}
	err = StoreKey(d, []byte(email), key)
	if err != nil {
		return nil, err
	}

	return []byte(key), nil
}

func StoreKey(d db.DB, email, pwhash []byte) error {
	return d.Update(storeKey(email, pwhash))
}

// TODO: Fixme!
func storeKey(email, pwhash []byte) func(*bolt.Tx) error {
	return func(*bolt.Tx) error {
		return nil
	}
}

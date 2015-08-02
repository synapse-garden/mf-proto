package testing

import (
	"os"

	"github.com/boltdb/bolt"
	"github.com/juju/errors"
	"github.com/synapse-garden/mf-proto/db"
)

// Fulfills db.DB
type TestingDB struct {
	bolt.DB
	filename    string
	updateError error
	viewError   error
}

type setupFunc func(*TestingDB) error

func NewTestingDB(setup ...setupFunc) (*TestingDB, error) {
	t := &TestingDB{}
	for _, f := range setup {
		err := f(t)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func CleanupDB(t *TestingDB) error {
	err := t.Close()
	if err != nil {
		return err
	}

	return os.Remove(t.filename)
}

func SetupBolt(name string) setupFunc {
	return func(t *TestingDB) error {
		var err error
		if f, err := os.Open(name); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		} else {
			if err = f.Close(); err != nil {
				return err
			}
			if err := os.Remove(name); err != nil {
				return err
			}
		}

		d, err := bolt.Open(name, 0600, nil)
		if err != nil {
			return err
		}
		t.filename = name
		t.DB = *d
		return nil
	}
}

func SetupBuckets(buckets ...[]db.Bucket) setupFunc {
	return func(t *TestingDB) error {
		for _, bs := range buckets {
			err := db.SetupBuckets(t, bs)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func SetupUpdateErr(msg string, vals ...interface{}) setupFunc {
	return func(t *TestingDB) error {
		t.updateError = errors.Errorf(msg, vals...)
		return nil
	}
}

func SetupViewErr(msg string, vals ...interface{}) setupFunc {
	return func(t *TestingDB) error {
		t.viewError = errors.Errorf(msg, vals...)
		return nil
	}
}

func (t *TestingDB) Update(fn func(*bolt.Tx) error) error {
	if t.updateError != nil {
		return t.updateError
	}
	return t.DB.Update(fn)
}

func (t *TestingDB) View(fn func(*bolt.Tx) error) error {
	if t.viewError != nil {
		return t.viewError
	}
	return t.DB.View(fn)
}

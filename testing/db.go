package testing

import (
	"os"

	"github.com/boltdb/bolt"
	"github.com/juju/errors"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/util"
)

// DB fulfills db.DB with attributes that can be manually set.
type DB struct {
	bolt.DB
	filename    string
	updateError error
	viewError   error
}

// SetupFunc is a func which sets up a *DB.
type SetupFunc func(*DB) error

// NewDB makes a new DB with the given setup func(s).
// Usage:
// import t "github.com/mf-proto/testing"
// db, err := t.NewDB(
//     t.SetupBolt("test.db"),
//     t.SetupBuckets("Admins", "Users"),
// )
func NewDB(setup ...SetupFunc) (*DB, error) {
	t := &DB{}
	for _, f := range setup {
		err := f(t)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

// CleanupDB closes and removes the DB.
func CleanupDB(t *DB) error {
	if err := t.Close(); err != nil {
		return err
	}

	return os.Remove(t.filename)
}

// SetupBolt sets up a BoltDB testing DB with the given filename.
func SetupBolt(name string) SetupFunc {
	return func(t *DB) error {
		if err := util.EnsureFileRemoved(name); err != nil {
			return err
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

// SetupBuckets adds the given db.Buckets to the DB.
func SetupBuckets(buckets ...[]db.Bucket) SetupFunc {
	return func(t *DB) error {
		for _, bs := range buckets {
			if err := db.SetupBuckets(t, bs); err != nil {
				return err
			}
		}
		return nil
	}
}

// SetupUpdateErr adds an error to be returned when Update is called for the DB.
func SetupUpdateErr(msg string, vals ...interface{}) SetupFunc {
	return func(t *DB) error {
		t.updateError = errors.Errorf(msg, vals...)
		return nil
	}
}

// SetupViewErr adds an error to be returned when View is called for the DB.
func SetupViewErr(msg string, vals ...interface{}) SetupFunc {
	return func(t *DB) error {
		t.viewError = errors.Errorf(msg, vals...)
		return nil
	}
}

// Update returns the pre-configured update error, or calls through to the
// underlying bolt DB.
func (t *DB) Update(fn func(*bolt.Tx) error) error {
	if t.updateError != nil {
		return t.updateError
	}
	return t.DB.Update(fn)
}

// View returns the pre-configured update error, or calls through to the
// underlying bolt DB.
func (t *DB) View(fn func(*bolt.Tx) error) error {
	if t.viewError != nil {
		return t.viewError
	}
	return t.DB.View(fn)
}

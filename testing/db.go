package testing

import (
	"fmt"

	"github.com/boltdb/bolt"
)

// Fulfills db.DB
type TestingDB struct {
	giveError error
	log       []string
}

type setupFunc func(t *TestingDB)

func NewTestingDB(setup ...setupFunc) *TestingDB {
	t := &TestingDB{nil, make([]string, 0)}
	for _, f := range setup {
		f(t)
	}
	return t
}

func Err(msg string, vals ...interface{}) setupFunc {
	return func(t *TestingDB) {
		t.giveError = fmt.Errorf(msg, vals...)
	}
}

func (t *TestingDB) Update(fn func(*bolt.Tx) error) error {
	t.log = append(t.log, "Update")
	return nil
}

func (t *TestingDB) View(fn func(*bolt.Tx) error) error {
	t.log = append(t.log, "View")
	return nil
}

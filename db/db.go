package db

import (
	"fmt"

	"github.com/boltdb/bolt"
)

type DB interface {
	// Begin(writable bool) (*bolt.Tx, error)
	// Close() error
	// GoString() string
	// Info() *Info
	// Path() string
	// Stats() Stats
	// String() string
	Update(fn func(*bolt.Tx) error) error
	// View(fn func(*bolt.Tx) error) error
}

func SetupBuckets(d DB, buckets []string) error {
	return d.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return fmt.Errorf("error creating bucket: %s", err)
			}
		}
		return nil
	})
}

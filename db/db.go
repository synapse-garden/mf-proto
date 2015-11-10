package db

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/juju/errors"
)

// DB specifies the basic methods that a database must implement to be used as
// a backend.
type DB interface {
	// Begin(writable bool) (*bolt.Tx, error)
	// Close() error
	// GoString() string
	// Info() *Info
	// Path() string
	// Stats() Stats
	// String() string
	Update(fn func(*bolt.Tx) error) error
	View(fn func(*bolt.Tx) error) error
}

// Bucket is a named database partition.
type Bucket string

// BucketNotFoundErr indicates that the given Bucket was not yet created.
func BucketNotFoundErr(b Bucket) error {
	return errors.NotFoundf("bucket %q not found", b)
}

// SetupBuckets creates the given Buckets if they do not already exist in d.
func SetupBuckets(d DB, buckets []Bucket) error {
	return d.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return errors.Annotatef(err, "error creating bucket %q", bucket)
			}
		}
		return nil
	})
}

// StoreKeyValue marshals the given value as JSON and stores it in d at the
// given key.
func StoreKeyValue(d DB, bucket Bucket, key []byte, value interface{}) error {
	vBytes, err := json.Marshal(value)
	if err != nil {
		return errors.Annotatef(err, "marshaling %#v into %q failed", value, bucket)
	}

	return d.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return BucketNotFoundErr(bucket)
		}
		if err := b.Put(key, vBytes); err != nil {
			return err
		}
		return nil
	})
}

// GetByKey retrieves the value stored in d with the given key.
func GetByKey(d DB, bucket Bucket, key []byte) ([]byte, error) {
	var result []byte

	err := d.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return BucketNotFoundErr(bucket)
		}
		result = b.Get(key)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteByKey deletes the value stored with the given key from d.  If key is
// not found, the error returned will be nil.
func DeleteByKey(d DB, bucket Bucket, key []byte) error {
	return d.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return BucketNotFoundErr(bucket)
		}
		return b.Delete(key)
	})
}

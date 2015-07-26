package db

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/juju/errors"
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
	View(fn func(*bolt.Tx) error) error
}

type Bucket string

func BucketNotFoundErr(b Bucket) error {
	return errors.NotFoundf("bucket %q not found", b)
}

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

func DeleteByKey(d DB, bucket Bucket, key []byte) error {
	return d.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return BucketNotFoundErr(bucket)
		}
		return b.Delete(key)
	})
}

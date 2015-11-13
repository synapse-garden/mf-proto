package object

import (
	"encoding/json"
	"fmt"

	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/util"

	errors "github.com/juju/errors"
)

const (
	// Objects is the bucket that contains all Objects.
	Objects db.Bucket = "object-objects"
)

// Buckets returns the Buckets for the object database.
func Buckets() []db.Bucket {
	return []db.Bucket{
		Objects,
	}
}

// Object is an object containing its own permissions.
type Object struct {
	// Json contains arbitrary text.
	JSON string `json:"json,omitempty"`

	// Perms defines the permissions for the object.
	Perms util.Permissions `json:"perms,omitempty"`
}

// New makes an object with the given json and default (owner only)
// permissions.
func New(json, user string) *Object {
	return &Object{
		JSON:  json,
		Perms: util.Permissions{user},
	}
}

// ReadAuthorized determines if a user is authorized to use an object.
func (o *Object) ReadAuthorized(email string) error {
	return o.Perms.ReadAuthorized(email)
}

// WriteAuthorized determines if a user is authorized to write an object.
func (o *Object) WriteAuthorized(email string) error {
	return o.Perms.WriteAuthorized(email)
}

// Put stores an object by id for the given user, if the user is authorized.
func Put(d db.DB, email string, id util.Key, obj *Object) error {
	o, err := Get(d, email, id)

	switch {
	case err != nil && errors.IsNotFound(err):
		return db.StoreKeyValue(d, Objects, []byte(id), obj)
	case err != nil && errors.IsUnauthorized(err):
		return errors.Annotatef(err,
			"user %q does not have read permissions for %s",
			email, id,
		)
	}

	if err = o.WriteAuthorized(email); err != nil {
		return errors.Annotatef(err,
			"user %q does not have write permissions for %s",
			email, id,
		)
	}

	return db.StoreKeyValue(d, Objects, []byte(id), obj)
}

// Get fetches an object by ID, if the user has permission to view it.
func Get(d db.DB, email string, id util.Key) (*Object, error) {
	objBytes, err := db.GetByKey(d, Objects, []byte(id))
	if err != nil {
		return nil, err
	}

	if len(objBytes) == 0 {
		return nil, errors.NotFoundf("object %s", id)
	}

	obj := new(Object)
	if err := json.Unmarshal(objBytes, obj); err != nil {
		return nil, errors.Annotatef(
			err, "unmarshaling %#q failed", objBytes,
		)
	}

	if err = obj.ReadAuthorized(email); err != nil {
		return nil, err
	}

	return obj, nil
}

// Delete deletes an object given a user and an Object id.
func Delete(d db.DB, email string, id util.Key) error {
	objBytes, err := db.GetByKey(d, Objects, []byte(id))
	if err != nil {
		return err
	}

	if len(objBytes) == 0 {
		// Was already deleted, no problem
		return nil
	}

	obj := new(Object)
	if err := json.Unmarshal(objBytes, obj); err != nil {
		return errors.Annotatef(
			err, "unmarshaling %#q failed", objBytes,
		)
	}

	if err = obj.ReadAuthorized(email); err != nil {
		return err
	}

	if err = obj.WriteAuthorized(email); err != nil {
		return err
	}

	return db.DeleteByKey(d, Objects, []byte(id))
}

// DeleteAll deletes all Objects owned by the given user.
func DeleteAll(d db.DB, email string) error {
	return fmt.Errorf("implement me")
}

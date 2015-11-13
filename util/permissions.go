package util

import "github.com/juju/errors"

// Permissions represents the permissions of an object.
type Permissions struct {
	Owner string
}

// ReadAuthorized determines if a given user is read authorized.
func (p *Permissions) ReadAuthorized(email string) error {
	if p.Owner != email {
		return errors.Unauthorizedf("user %q not read authorized", email)
	}

	return nil
}

// WriteAuthorized determines if a given user is write authorized.
func (p *Permissions) WriteAuthorized(email string) error {
	if p.Owner != email {
		return errors.Unauthorizedf("user %q not write authorized", email)
	}

	return nil
}

package util

import "github.com/juju/errors"

// Permissions represents the permissions of an object.
type Permissions struct {
	Owner string
}

// Authorized determines if a given user is authorized.
func (p *Permissions) Authorized(email string) error {
	if p.Owner != email {
		return errors.Unauthorizedf("user %q not authorized", email)
	}

	return nil
}

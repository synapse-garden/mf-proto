package testing

import (
	"time"

	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/user"
	"github.com/synapse-garden/mf-proto/util"
)

type TestUser struct {
	Email     string
	Pwhash    string
	LoginKey  string
	LoginHash string
}

type TestAdmin struct {
	Email  string
	Pwhash string
	Key    util.Key
}

func LoginUsers(d db.DB, users ...TestUser) SetupFunc {
	var err error
	user.SetTimeout(5 * time.Second)
	return func(t *DB) error {
		for _, u := range users {
			err = db.StoreKeyValue(
				t,
				user.LoginKeys,
				[]byte(u.Email),
				user.Login{
					util.Key(u.LoginKey),
					time.Now().Add(user.GetTimeout()),
				},
			)

			if err != nil {
				return err
			}
		}

		return err
	}
}

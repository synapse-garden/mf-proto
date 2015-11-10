package testing

import (
	"github.com/juju/errors"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/object"
	"github.com/synapse-garden/mf-proto/util"
)

// CreateObjects stores the given map of IDs to Objects in the database.
func CreateObjects(d db.DB, obs map[util.Key]*object.Object) {
	for k, o := range obs {
		err := db.StoreKeyValue(d, object.Objects, []byte(k), o)
		if err != nil {
			panic(err)
		}
	}
}

// CleanupObjects cleans the given map of IDs to Objects in the database.
func CleanupObjects(d db.DB, obs []util.Key) {
	for _, k := range obs {
		err := db.DeleteByKey(d, object.Objects, []byte(k))
		switch {
		case err != nil && errors.IsNotFound(err):
		case err != nil:
			panic(err)
		}
	}
}

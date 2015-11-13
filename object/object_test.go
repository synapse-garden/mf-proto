package object_test

import (
	"encoding/json"
	"testing"

	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/object"
	mft "github.com/synapse-garden/mf-proto/testing"
	"github.com/synapse-garden/mf-proto/util"

	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type ObjectSuite struct {
	d       *mft.DB
	objects map[util.Key]object.Object
}

var _ = gc.Suite(&ObjectSuite{})

func (s *ObjectSuite) SetUpTest(c *gc.C) {
	d, err := mft.NewDB(
		mft.SetupBolt("test.db"),
		mft.SetupBuckets(object.Buckets()),
	)
	c.Assert(err, jc.ErrorIsNil)
	s.d = d
}

func (s *ObjectSuite) TearDownTest(c *gc.C) {
	if d := s.d; d != nil {
		c.Assert(mft.CleanupDB(d), jc.ErrorIsNil)
	}
}

func (s *ObjectSuite) TestReadAuthorized(c *gc.C) {
	for i, t := range []struct {
		should             string
		given              *object.Object
		givenEmail         string
		expectError        string
		expectUnauthorized bool
	}{{
		should: "reject an unauthorized user",
		given: &object.Object{
			Perms: util.Permissions{"joe"},
		},
		givenEmail:         "not-joe",
		expectError:        `user "not-joe" not read authorized`,
		expectUnauthorized: true,
	}, {
		should:     "accept an authorized user",
		given:      &object.Object{Perms: util.Permissions{"joe"}},
		givenEmail: "joe",
	}} {
		c.Logf("test %d: should %s", i, t.should)
		err := t.given.ReadAuthorized(t.givenEmail)
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			c.Check(errors.IsUnauthorized(err), gc.Equals, t.expectUnauthorized)
		} else {
			c.Check(err, jc.ErrorIsNil)
		}
	}
}

func (s *ObjectSuite) TestWriteAuthorized(c *gc.C) {
	for i, t := range []struct {
		should             string
		given              *object.Object
		givenEmail         string
		expectError        string
		expectUnauthorized bool
	}{{
		should: "reject an unauthorized user",
		given: &object.Object{
			Perms: util.Permissions{"joe"},
		},
		givenEmail:         "not-joe",
		expectError:        `user "not-joe" not write authorized`,
		expectUnauthorized: true,
	}, {
		should:     "accept an authorized user",
		given:      &object.Object{Perms: util.Permissions{"joe"}},
		givenEmail: "joe",
	}} {
		c.Logf("test %d: should %s", i, t.should)
		err := t.given.WriteAuthorized(t.givenEmail)
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			c.Check(errors.IsUnauthorized(err), gc.Equals, t.expectUnauthorized)
		} else {
			c.Check(err, jc.ErrorIsNil)
		}
	}
}

func (s *ObjectSuite) TestPut(c *gc.C) {
	for i, t := range []struct {
		should               string
		givenExistingObjects map[util.Key]*object.Object
		givenNewObject       *object.Object
		givenUser            string
		givenID              string
		expectError          string
		expectUnauthorized   bool
	}{{
		should:         "put a new object for a valid user",
		givenNewObject: object.New("foo", "joe"),
		givenUser:      "joe",
		givenID:        "12345",
	}, {
		should: "overwrite an object for its owner",
		givenExistingObjects: map[util.Key]*object.Object{
			"12345": object.New("foo", "joe"),
		},
		givenUser:      "joe",
		givenNewObject: object.New("bar", "joe"),
		givenID:        "12345",
	}, {
		should: "not overwrite an object for new user",
		givenExistingObjects: map[util.Key]*object.Object{
			"12345": object.New("foo", "joe"),
		},
		givenUser:          "fred",
		givenNewObject:     object.New("bar", "fred"),
		givenID:            "12345",
		expectError:        `user "fred" does not have read permissions for 12345: user "fred" not read authorized`,
		expectUnauthorized: true,
	}, {
		should:             "not put a new object for an invalid user",
		givenNewObject:     object.New("foo", "joe"),
		givenUser:          "not-joe",
		givenID:            "12345",
		expectError:        `user "not-joe" does not have read permissions for 12345: user "not-joe" not read authorized`,
		expectUnauthorized: true,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		mft.CreateObjects(s.d, t.givenExistingObjects)

		err := object.Put(s.d, t.givenUser, util.Key(t.givenID), t.givenNewObject)

		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			c.Check(
				errors.IsUnauthorized(err),
				gc.Equals,
				t.expectUnauthorized,
			)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)

		obj, err := object.Get(s.d, t.givenUser, util.Key(t.givenID))
		c.Assert(err, jc.ErrorIsNil)
		c.Check(obj, jc.DeepEquals, t.givenNewObject)
	}
}

func (s *ObjectSuite) TestGet(c *gc.C) {
	for i, t := range []struct {
		should               string
		givenExistingObjects map[util.Key]*object.Object
		givenUser            string
		givenID              string
		expectObject         *object.Object
		expectError          string
		expectUnauthorized   bool
	}{{
		should: "get an object for a valid user",
		givenExistingObjects: map[util.Key]*object.Object{
			"12345": object.New("foo", "joe"),
		},
		givenUser:    "joe",
		givenID:      "12345",
		expectObject: object.New("foo", "joe"),
	}, {
		should: "not get an object if user has no read permissions",
		givenExistingObjects: map[util.Key]*object.Object{
			"12345": object.New("foo", "joe"),
		},
		givenUser:          "fred",
		givenID:            "12345",
		expectError:        `user "fred" not read authorized`,
		expectUnauthorized: true,
	}, {
		should: "not get an object which does not exist",
		givenExistingObjects: map[util.Key]*object.Object{
			"12345": object.New("foo", "joe"),
		},
		givenUser:   "joe",
		givenID:     "123456",
		expectError: "object 123456 not found",
	}} {
		c.Logf("test %d: should %s", i, t.should)

		mft.CreateObjects(s.d, t.givenExistingObjects)

		obj, err := object.Get(s.d, t.givenUser, util.Key(t.givenID))

		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			c.Check(
				errors.IsUnauthorized(err),
				gc.Equals,
				t.expectUnauthorized,
			)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)
		c.Check(obj, jc.DeepEquals, t.expectObject)
	}
}

func (s *ObjectSuite) TestDelete(c *gc.C) {
	for i, t := range []struct {
		should               string
		givenExistingObjects map[util.Key]*object.Object
		givenUser            string
		givenID              string
		expectError          string
		expectUnauthorized   bool
		expectObjects        map[util.Key]*object.Object
	}{{
		should: "delete an object for a valid user",
		givenExistingObjects: map[util.Key]*object.Object{
			"12345":  object.New("foo", "joe"),
			"123456": object.New("foo", "joe"),
		},
		givenUser: "joe",
		givenID:   "12345",
	}, {
		should: "not overwrite an object for someone else",
		givenExistingObjects: map[util.Key]*object.Object{
			"12345": object.New("foo", "joe"),
		},
		givenUser:          "fred",
		givenID:            "12345",
		expectError:        `user "fred" not read authorized`,
		expectUnauthorized: true,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		mft.CreateObjects(s.d, t.givenExistingObjects)

		err := object.Delete(s.d, t.givenUser, util.Key(t.givenID))

		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			c.Check(
				errors.IsUnauthorized(err),
				gc.Equals,
				t.expectUnauthorized,
			)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)

		// Make a map of expected objects minus the deleted one.
		expectObjects := make(map[util.Key]*object.Object)
		for id, obj := range t.givenExistingObjects {
			expectObjects[id] = obj
		}

		delete(expectObjects, util.Key(t.givenID))

		// Check the original objects to make sure the given ID was
		// deleted.
		for id, obj := range t.givenExistingObjects {
			objBytes, err := db.GetByKey(s.d, object.Objects, []byte(id))
			c.Assert(err, jc.ErrorIsNil)
			if id == util.Key(t.givenID) {
				// Make sure the given ID got deleted and is
				// not found.
				c.Check(len(objBytes), gc.Equals, 0)
				continue
			}

			c.Assert(len(objBytes), gc.Not(gc.Equals), 0)
			o := new(object.Object)
			c.Assert(json.Unmarshal(objBytes, o), jc.ErrorIsNil)
			c.Check(o, jc.DeepEquals, obj)
		}
	}
}

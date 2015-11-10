package object_test

import (
	"testing"

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

func (s *ObjectSuite) TestAuthorized(c *gc.C) {
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
		expectError:        `user "not-joe" not authorized`,
		expectUnauthorized: true,
	}, {
		should:     "accept an authorized user",
		given:      &object.Object{Perms: util.Permissions{"joe"}},
		givenEmail: "joe",
	}} {
		c.Logf("test %d: should %s", i, t.should)
		err := t.given.Authorized(t.givenEmail)
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
		should:             "not put a new object for an invalid user",
		givenNewObject:     object.New("foo", "joe"),
		givenUser:          "not-joe",
		givenID:            "12345",
		expectError:        `user "not-joe" not authorized`,
		expectUnauthorized: true,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		mft.CreateObjects(s.d, t.givenExistingObjects)

		err := object.Put(s.d, t.givenUser, util.Key(t.givenID), t.givenNewObject)

		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			c.Check(errors.IsUnauthorized(err), gc.Equals, t.expectUnauthorized)
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
		should string
	}{{}} {
		c.Logf("test %d: should %s", i, t.should)
	}
}

func (s *ObjectSuite) TestDelete(c *gc.C) {

}

package util_test

import (
	jc "github.com/juju/testing/checkers"
	"github.com/synapse-garden/mf-proto/util"
	gc "gopkg.in/check.v1"
)

func (s *UtilSuite) TestReadAuthorized(c *gc.C) {
	for i, t := range []struct {
		should      string
		givenPerms  util.Permissions
		givenEmail  string
		expectError string
	}{{
		should:     "accept read for an authorized user",
		givenPerms: util.Permissions{"joe"},
		givenEmail: "joe",
	}, {
		should:      "reject read for an unauthorized user",
		givenPerms:  util.Permissions{"joe"},
		givenEmail:  "fred",
		expectError: `user "fred" not read authorized`,
	}} {
		c.Logf("test %d: should %s", i, t.should)
		err := t.givenPerms.ReadAuthorized(t.givenEmail)
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
		} else {
			c.Check(err, jc.ErrorIsNil)
		}
	}
}

func (s *UtilSuite) TestWriteAuthorized(c *gc.C) {
	for i, t := range []struct {
		should      string
		givenPerms  util.Permissions
		givenEmail  string
		expectError string
	}{{
		should:     "accept write for an authorized user",
		givenPerms: util.Permissions{"joe"},
		givenEmail: "joe",
	}, {
		should:      "reject write for an unauthorized user",
		givenPerms:  util.Permissions{"joe"},
		givenEmail:  "fred",
		expectError: `user "fred" not write authorized`,
	}} {
		c.Logf("test %d: should %s", i, t.should)
		err := t.givenPerms.WriteAuthorized(t.givenEmail)
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
		} else {
			c.Check(err, jc.ErrorIsNil)
		}
	}
}

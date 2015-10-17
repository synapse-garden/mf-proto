package user_test

import (
	"encoding/json"
	"fmt"

	jc "github.com/juju/testing/checkers"
	"github.com/synapse-garden/mf-proto/db"
	t "github.com/synapse-garden/mf-proto/testing"
	"github.com/synapse-garden/mf-proto/user"
	"github.com/synapse-garden/mf-proto/util"

	gc "gopkg.in/check.v1"
)

func (s *UserSuite) TestCreate(c *gc.C) {
	for i, t := range []struct {
		should      string
		user        t.TestUser
		expectError string
	}{{
		should: "create a user",
		user:   s.users["bob"],
	}, {
		should:      "not create an existing user",
		user:        s.users["bob"],
		expectError: `user for email "bob@tomato.com" already exists`,
	}, {
		should: "create a new different user",
		user:   s.users["larry"],
	}, {
		should:      "not create the same user twice",
		user:        s.users["larry"],
		expectError: `user for email "larry@cucumber.net" already exists`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		if t.expectError == "" {
			c.Check(s.createTests(t.user, c), jc.ErrorIsNil)
		} else {
			c.Check(s.createTests(t.user, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *UserSuite) createTests(u t.TestUser, c *gc.C) error {
	err := user.Create(s.d, u.Email, u.Pwhash)
	if err != nil {
		return err
	}

	userBytes, err := db.GetByKey(s.d, user.Users, []byte(u.Email))
	if err != nil {
		return err
	}
	c.Assert(len(userBytes), gc.Not(gc.Equals), 0)

	var tmpUser user.User
	err = json.Unmarshal(userBytes, &tmpUser)
	if err != nil {
		return err
	}

	c.Check(tmpUser.Email, gc.Equals, u.Email)
	return nil
}

func (s *UserSuite) TestDelete(c *gc.C) {
	s.createUsers(c)

	for i, t := range []struct {
		should      string
		user        t.TestUser
		expectError string
	}{{
		should: "delete a user",
		user:   s.users["bob"],
	}, {
		should:      "not delete a nonexistent user",
		user:        t.TestUser{Email: "jove@olympus.mons", LoginKey: "foo"},
		expectError: `user for email "jove@olympus.mons" not found`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		if t.expectError == "" {
			c.Check(s.deleteTests(t.user, c), jc.ErrorIsNil)
		} else {
			c.Check(s.deleteTests(t.user, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *UserSuite) deleteTests(u t.TestUser, c *gc.C) error {
	if err := user.Delete(s.d, u.Email); err != nil {
		return err
	}

	err := user.ValidLogin(s.d, u.Email, util.Key(u.LoginKey))
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf("could not get login for email %q not valid", u.Email))
	return nil
}

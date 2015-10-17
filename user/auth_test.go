package user_test

import (
	"encoding/json"
	"fmt"
	"time"

	jc "github.com/juju/testing/checkers"
	"github.com/synapse-garden/mf-proto/db"
	t "github.com/synapse-garden/mf-proto/testing"
	"github.com/synapse-garden/mf-proto/user"
	"github.com/synapse-garden/mf-proto/util"

	gc "gopkg.in/check.v1"
)

func (s *UserSuite) TestLoginUser(c *gc.C) {
	for i, t := range []struct {
		should      string
		user        t.TestUser
		expectError string
	}{{
		should: "be able to log in an existing user",
		user:   s.users["bob"],
	}, {
		should:      "not log in a nonexistent user",
		user:        t.TestUser{Email: "jove@olympus.mons", Pwhash: "1000"},
		expectError: `"jove@olympus.mons" user not found`,
	}} {
		c.Logf("test %d: should %s", i, t.should)
		if t.expectError == "" {
			c.Check(s.testLogins(t.user, c), jc.ErrorIsNil)
		} else {
			c.Check(s.testLogins(t.user, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *UserSuite) testLogins(u t.TestUser, c *gc.C) error {
	s.createUsers(c)
	defer s.deleteUsers(c)

	key, err := user.LoginUser(s.d, u.Email, u.Pwhash)
	if err != nil {
		return err
	}

	userBytes, err := db.GetByKey(s.d, user.LoginKeys, []byte(u.Email))
	if err != nil {
		return err
	}

	c.Assert(len(userBytes), gc.Not(gc.Equals), 0)

	var login user.Login
	err = json.Unmarshal(userBytes, &login)
	if err != nil {
		return err
	}

	c.Assert(login.Key, gc.Equals, key)
	return nil
}

func (s *UserSuite) TestValidUser(c *gc.C) {
	s.createUsers(c)
	b := s.users["bob"]
	key, err := user.LoginUser(s.d, b.Email, b.Pwhash)
	c.Assert(err, jc.ErrorIsNil)
	s.users["bob"] = t.TestUser{
		Email:    b.Email,
		LoginKey: string(key),
	}

	for i, t := range []struct {
		should      string
		user        t.TestUser
		pause       int
		expectError string
	}{{
		should: "validate an existing login",
		user:   s.users["bob"],
	}, {
		should:      "not validate an existing login with bad key",
		user:        t.TestUser{Email: b.Email, LoginKey: "foo"},
		expectError: `bad key "foo" for user "bob@tomato.com" not valid`,
	}, {
		should:      "not validate a nonexistent user",
		user:        t.TestUser{Email: "jove@olympus.mons", LoginKey: "foo"},
		expectError: `could not get login for email "jove@olympus.mons" not valid`,
	}, {
		should:      "not validate a timed-out user",
		pause:       80,
		user:        s.users["bob"],
		expectError: `user "bob@tomato.com" timed out not valid`,
	}} {
		c.Logf("test %d: should %s", i, t.should)
		if t.expectError == "" {
			c.Check(s.testValidate(t.user, t.pause, c), jc.ErrorIsNil)
		} else {
			c.Check(s.testValidate(t.user, t.pause, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *UserSuite) testValidate(u t.TestUser, pause int, c *gc.C) error {
	<-time.NewTimer(timeout(pause)).C
	err := user.ValidLogin(s.d, u.Email, util.Key(u.LoginKey))
	if err != nil {
		return err
	}

	return nil
}

func (s *UserSuite) TestLogoutUser(c *gc.C) {
	s.createUsers(c)

	for i, t := range []struct {
		should      string
		login       bool
		user        t.TestUser
		key         util.Key
		expectError string
	}{{
		should: "log out a logged-in user",
		login:  true,
		user:   s.users["bob"],
	}, {
		should:      "not logout a logged-out user",
		user:        s.users["larry"],
		key:         "12345",
		expectError: `could not get login for email "larry@cucumber.net" not valid`,
	}, {
		should:      "not logout a nonexistent user",
		user:        t.TestUser{Email: "jove@olympus.mons"},
		key:         "12345",
		expectError: `could not get login for email "jove@olympus.mons" not valid`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		if t.expectError == "" {
			c.Check(s.logoutTests(t.user, t.key, t.login, c), jc.ErrorIsNil)
		} else {
			c.Check(s.logoutTests(t.user, t.key, t.login, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *UserSuite) logoutTests(u t.TestUser, key util.Key, login bool, c *gc.C) error {
	if login {
		key, err := user.LoginUser(s.d, u.Email, u.Pwhash)
		c.Assert(err, jc.ErrorIsNil)
		u = t.TestUser{
			Email:    u.Email,
			Pwhash:   u.Pwhash,
			LoginKey: string(key),
		}
	}

	if key != util.Key("") {
		u.LoginKey = string(key)
	}

	err := user.LogoutUser(s.d, u.Email, util.Key(u.LoginKey))
	if err != nil {
		return err
	}

	err = user.LogoutUser(s.d, u.Email, util.Key(u.LoginKey))
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf("could not get login for email %q not valid", u.Email))
	return nil
}

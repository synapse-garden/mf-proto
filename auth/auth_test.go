package auth_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/synapse-garden/mf-proto/auth"
	"github.com/synapse-garden/mf-proto/db"
	t "github.com/synapse-garden/mf-proto/testing"
	"github.com/synapse-garden/mf-proto/util"
	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type AuthSuite struct {
	d     *t.TestingDB
	users map[string]t.TestUser
}

var _ = gc.Suite(&AuthSuite{})

func timeout(t int) time.Duration {
	return time.Duration(t) * time.Millisecond
}

func (s *AuthSuite) SetUpTest(c *gc.C) {
	d, err := t.NewTestingDB(
		t.SetupBolt("test.db"),
		t.SetupBuckets(auth.Buckets()),
	)
	c.Assert(err, gc.IsNil)
	s.d = d
	s.users = map[string]t.TestUser{
		"bob": {
			Email:  "bob@tomato.com",
			Pwhash: "12345",
		},
		"larry": {
			Email:  "larry@cucumber.net",
			Pwhash: "54321",
		},
	}
	auth.SetTimeout(time.Duration(50) * time.Millisecond)
}

func (s *AuthSuite) TearDownTest(c *gc.C) {
	s.users = nil
	c.Assert(t.CleanupDB(s.d), gc.IsNil)
	auth.SetTimeout(time.Duration(5) * time.Minute)
}

func (s *AuthSuite) createUsers(c *gc.C) {
	for _, u := range s.users {
		err := auth.CreateUser(s.d, u.Email, u.Pwhash)
		c.Assert(err, gc.IsNil)
	}
}

func (s *AuthSuite) deleteUsers(c *gc.C) {
	for _, u := range s.users {
		err := auth.DeleteUser(s.d, u.Email)
		c.Assert(err, gc.IsNil)
	}
}

func (s *AuthSuite) TestCreateUser(c *gc.C) {
	for i, t := range []struct {
		should      string
		user        t.TestUser
		expectError string
	}{{
		should: "Create a user",
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
			c.Check(s.createTests(t.user, c), gc.IsNil)
		} else {
			c.Check(s.createTests(t.user, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *AuthSuite) createTests(user t.TestUser, c *gc.C) error {
	err := auth.CreateUser(s.d, user.Email, user.Pwhash)
	if err != nil {
		return err
	}

	userBytes, err := db.GetByKey(s.d, auth.Users, []byte(user.Email))
	if err != nil {
		return err
	}
	c.Assert(len(userBytes), gc.Not(gc.Equals), 0)

	var tmpUser auth.User
	err = json.Unmarshal(userBytes, &tmpUser)
	if err != nil {
		return err
	}

	c.Check(tmpUser.Email, gc.Equals, user.Email)
	return nil
}

func (s *AuthSuite) TestLoginUser(c *gc.C) {
	for i, t := range []struct {
		should      string
		user        t.TestUser
		expectError string
	}{{
		should: "Be able to log in an existing user",
		user:   s.users["bob"],
	}, {
		should:      "Not log in a nonexistent user",
		user:        t.TestUser{Email: "jove@olympus.mons", Pwhash: "1000"},
		expectError: `no user for email "jove@olympus.mons"`,
	}} {
		c.Logf("test %d: should %s", i, t.should)
		if t.expectError == "" {
			c.Check(s.testLogins(t.user, c), gc.IsNil)
		} else {
			c.Check(s.testLogins(t.user, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *AuthSuite) testLogins(user t.TestUser, c *gc.C) error {
	s.createUsers(c)
	defer s.deleteUsers(c)

	key, err := auth.LoginUser(s.d, user.Email, user.Pwhash)
	if err != nil {
		return err
	}

	userBytes, err := db.GetByKey(s.d, auth.SessionKeys, []byte(user.Email))
	if err != nil {
		return err
	}

	c.Assert(len(userBytes), gc.Not(gc.Equals), 0)

	var login auth.Login
	err = json.Unmarshal(userBytes, &login)
	if err != nil {
		return err
	}

	c.Assert(login.Key, gc.Equals, key)
	return nil
}

func (s *AuthSuite) TestValidUser(c *gc.C) {
	s.createUsers(c)
	b := s.users["bob"]
	key, err := auth.LoginUser(s.d, b.Email, b.Pwhash)
	c.Assert(err, gc.IsNil)
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
		should: "Validate an existing login",
		user:   s.users["bob"],
	}, {
		should:      "Not validate an existing login with bad key",
		user:        t.TestUser{Email: b.Email, LoginKey: "foo"},
		expectError: `bad key "foo" for user "bob@tomato.com"`,
	}, {
		should:      "Not validate a nonexistent user",
		user:        t.TestUser{Email: "jove@olympus.mons", LoginKey: "foo"},
		expectError: `user "jove@olympus.mons" not logged in user not found`,
	}, {
		should:      "Not validate a timed-out user",
		pause:       80,
		user:        s.users["bob"],
		expectError: `user "bob@tomato.com" timed out not valid`,
	}} {
		c.Logf("test %d: should %s", i, t.should)
		if t.expectError == "" {
			c.Check(s.testValidate(t.user, t.pause, c), gc.IsNil)
		} else {
			c.Check(s.testValidate(t.user, t.pause, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *AuthSuite) testValidate(user t.TestUser, pause int, c *gc.C) error {
	<-time.NewTimer(timeout(pause)).C
	err := auth.Valid(s.d, user.Email, util.Key(user.LoginKey))
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthSuite) TestDeleteUser(c *gc.C) {
	s.createUsers(c)

	for i, t := range []struct {
		should      string
		user        t.TestUser
		expectError string
	}{{
		should: "Delete a user",
		user:   s.users["bob"],
	}, {
		should:      "Not delete a nonexistent user",
		user:        t.TestUser{Email: "jove@olympus.mons", LoginKey: "foo"},
		expectError: `user for email "jove@olympus.mons" not found`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		if t.expectError == "" {
			c.Check(s.deleteTests(t.user, c), gc.IsNil)
		} else {
			c.Check(s.deleteTests(t.user, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *AuthSuite) deleteTests(user t.TestUser, c *gc.C) error {
	err := auth.DeleteUser(s.d, user.Email)
	if err != nil {
		return err
	}

	err = auth.Valid(s.d, user.Email, util.Key(user.LoginKey))
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf("user %q not logged in user not found", user.Email))
	return nil
}

func (s *AuthSuite) TestLogoutUser(c *gc.C) {
	s.createUsers(c)

	for i, t := range []struct {
		should      string
		login       bool
		user        t.TestUser
		key         util.Key
		expectError string
	}{{
		should: "Log out a logged-in user",
		login:  true,
		user:   s.users["bob"],
	}, {
		should:      "Not logout a logged-out user",
		user:        s.users["larry"],
		key:         "12345",
		expectError: `user "larry@cucumber.net" not logged in user not found`,
	}, {
		should:      "Not logout a nonexistent user",
		user:        t.TestUser{Email: "jove@olympus.mons"},
		key:         "12345",
		expectError: `user "jove@olympus.mons" not logged in user not found`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		if t.expectError == "" {
			c.Check(s.logoutTests(t.user, t.key, t.login, c), gc.IsNil)
		} else {
			c.Check(s.logoutTests(t.user, t.key, t.login, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *AuthSuite) logoutTests(user t.TestUser, key util.Key, login bool, c *gc.C) error {
	if login {
		key, err := auth.LoginUser(s.d, user.Email, user.Pwhash)
		c.Assert(err, gc.IsNil)
		user = t.TestUser{
			Email:    user.Email,
			Pwhash:   user.Pwhash,
			LoginKey: string(key),
		}
	}

	if key != util.Key("") {
		user.LoginKey = string(key)
	}

	err := auth.LogoutUser(s.d, user.Email, util.Key(user.LoginKey))
	if err != nil {
		return err
	}

	err = auth.LogoutUser(s.d, user.Email, util.Key(user.LoginKey))
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf("user %q not logged in user not found", user.Email))
	return nil
}

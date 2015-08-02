package user_test

import (
	"testing"
	"time"

	jc "github.com/juju/testing/checkers"
	t "github.com/synapse-garden/mf-proto/testing"
	"github.com/synapse-garden/mf-proto/user"

	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type UserSuite struct {
	d     *t.TestingDB
	users map[string]t.TestUser
}

var _ = gc.Suite(&UserSuite{})

func timeout(t int) time.Duration {
	return time.Duration(t) * time.Millisecond
}

func (s *UserSuite) SetUpTest(c *gc.C) {
	d, err := t.NewTestingDB(
		t.SetupBolt("test.db"),
		t.SetupBuckets(user.Buckets()),
	)
	c.Assert(err, jc.ErrorIsNil)
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
	user.SetTimeout(time.Duration(50) * time.Millisecond)
}

func (s *UserSuite) TearDownTest(c *gc.C) {
	s.users = nil
	c.Assert(t.CleanupDB(s.d), jc.ErrorIsNil)
	user.SetTimeout(time.Duration(5) * time.Minute)
}

func (s *UserSuite) createUsers(c *gc.C) {
	for _, u := range s.users {
		err := user.Create(s.d, u.Email, u.Pwhash)
		c.Assert(err, jc.ErrorIsNil)
	}
}

func (s *UserSuite) deleteUsers(c *gc.C) {
	for _, u := range s.users {
		err := user.Delete(s.d, u.Email)
		c.Assert(err, jc.ErrorIsNil)
	}
}

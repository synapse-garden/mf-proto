package auth_test

import (
	"testing"

	"github.com/synapse-garden/mf-proto/auth"
	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type AuthSuite struct {
	preparedUsers []auth.User
	createdUsers  []auth.User
	logins        []auth.Login
}

var _ = gc.Suite(&AuthSuite{})

func (s *AuthSuite) SetUpSuite(c *gc.C) {

}

func (s *AuthSuite) TestCreateUser(c *gc.C) {
	for i, t := range []struct {
		should        string
		givenEmail    string
		givenHashedPw string
		expect        string
	}{{
		should: "Create a few users",
	}} {
		c.Logf("test %d: should %s", i, t.should)
	}
}

func (s *AuthSuite) TestLoginUser(c *gc.C) {
	for i, t := range []struct {
		should        string
		givenEmail    string
		givenHashedPw string
		expect        string
	}{{
		should: "Create a few users",
	}} {
		c.Logf("test %d: should %s", i, t.should)
	}
}

package auth_test

import (
	"testing"

	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type AuthSuite struct{}

var _ = gc.Suite(&AuthSuite{})

func (s *AuthSuite) TestCreateUser(c *gc.C) {
	for i, t := range []struct {
		should        string
		givenEmail    string
		givenHashedPw string
		expect        string
	}{{}} {
		c.Logf("test %d: should %s", i, t.should)
	}
}

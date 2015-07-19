package util_test

import (
	"testing"

	"github.com/synapse-garden/mf-proto/util"
	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type UtilSuite struct{}

var _ = gc.Suite(&UtilSuite{})

func (s *UtilSuite) TestCheckHashedPw(c *gc.C) {
	for i, t := range []struct {
		should    string
		givenPw   string
		givenSalt util.Salt
		givenHash util.Hash
		expect    bool
	}{{
		should:    "work",
		givenPw:   "foobar",
		givenSalt: "c9fd228aa912e8a3f591590e486719af283598f0",
		givenHash: "edd40ea1fef74898d639b6cdce7610c518487e2a",
		expect:    true,
	}, {
		should:    "also work",
		givenPw:   "deadbeef",
		givenSalt: "125b43964f67f88d7de538b1d310c479822a5d0d",
		givenHash: "50aa2ddda4f15d637585d2843242cba76d130afc",
		expect:    true,
	}} {
		c.Logf("test %d: should %s", i, t.should)
		result := util.CheckHashedPw(t.givenPw, t.givenSalt, t.givenHash)
		c.Check(result, gc.Equals, t.expect)
	}
}

func (s *UtilSuite) TestHashedAndSalt(c *gc.C) {
	for i, t := range []struct {
		should     string
		givenPw    string
		givenSeed  string
		expectHash util.Hash
		expectSalt util.Salt
	}{{
		should:     "work",
		givenPw:    "foobar",
		givenSeed:  "seedFooBar",
		expectHash: "edd40ea1fef74898d639b6cdce7610c518487e2a",
		expectSalt: "c9fd228aa912e8a3f591590e486719af283598f0",
	}, {
		should:     "work",
		givenPw:    "deadbeef",
		givenSeed:  "anotherseed",
		expectHash: "50aa2ddda4f15d637585d2843242cba76d130afc",
		expectSalt: "125b43964f67f88d7de538b1d310c479822a5d0d",
	}} {
		c.Logf("test %d: should %s", i, t.should)
		h, s := util.HashedAndSalt(t.givenPw, t.givenSeed)
		c.Check(string(h), gc.Equals, string(t.expectHash))
		c.Check(string(s), gc.Equals, string(t.expectSalt))
	}
}

func (s *UtilSuite) TestSaltedHash(c *gc.C) {
	for i, t := range []struct {
		should    string
		givenPw   string
		givenSeed string
		expectKey util.Key
	}{{
		should:    "work",
		givenPw:   "foobar",
		givenSeed: "seedFooBar",
		expectKey: "edd40ea1fef74898d639b6cdce7610c518487e2a",
	}, {
		should:    "work",
		givenPw:   "deadbeef",
		givenSeed: "anotherseed",
		expectKey: "50aa2ddda4f15d637585d2843242cba76d130afc",
	}} {
		c.Logf("test %d: should %s", i, t.should)
		h := util.SaltedHash(t.givenPw, t.givenSeed)
		c.Check(string(h), gc.Equals, string(t.expectKey))
	}
}

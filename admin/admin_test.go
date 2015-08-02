package admin_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/synapse-garden/mf-proto/admin"
	"github.com/synapse-garden/mf-proto/db"
	t "github.com/synapse-garden/mf-proto/testing"
	"github.com/synapse-garden/mf-proto/util"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type AdminSuite struct {
	d      db.DB
	admins map[string]t.TestAdmin
}

var _ = gc.Suite(&AdminSuite{})

func (s *AdminSuite) SetUpTest(c *gc.C) {
	d, err := t.NewTestingDB(
		t.SetupBolt("test.db"),
		t.SetupBuckets(admin.Buckets()),
	)
	c.Assert(err, jc.ErrorIsNil)
	s.d = d

	s.admins = map[string]t.TestAdmin{
		"bob": {
			Email:  "bob@tomato.com",
			Pwhash: "12345",
		},
		"larry": {
			Email:  "larry@cucumber.net",
			Pwhash: "54321",
		},
	}
}

func (s *AdminSuite) createAdmins(c *gc.C) {
	for name, u := range s.admins {
		key, err := admin.Create(s.d, u.Email, u.Pwhash)
		c.Assert(err, jc.ErrorIsNil)
		u.Key = key
		s.admins[name] = u
	}
}

func (s *AdminSuite) deleteAdmins(c *gc.C) {
	for _, u := range s.admins {
		err := admin.Delete(s.d, u.Key)
		c.Assert(err, jc.ErrorIsNil)
	}
}

func (s *AdminSuite) TearDownTest(c *gc.C) {

}

func (s *AdminSuite) TestCreate(c *gc.C) {
	for i, t := range []struct {
		should      string
		admin       t.TestAdmin
		expectError string
	}{{
		should: "create a new admin",
		admin:  s.admins["bob"],
	}, {
		should:      "not create an existing admin",
		admin:       s.admins["bob"],
		expectError: `admin for email bob@tomato.com: already exists`,
	}, {
		should: "create a new different admin",
		admin:  s.admins["larry"],
	}, {
		should:      "not create the same admin twice",
		admin:       s.admins["larry"],
		expectError: `admin for email larry@cucumber.net: already exists`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		if t.expectError == "" {
			c.Check(s.createTests(t.admin, c), jc.ErrorIsNil)
		} else {
			c.Check(s.createTests(t.admin, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *AdminSuite) createTests(adm t.TestAdmin, c *gc.C) error {
	key, err := admin.Create(s.d, adm.Email, adm.Pwhash)
	if err != nil {
		return err
	}

	err = admin.IsAdmin(s.d, util.Key(key))
	c.Assert(err, jc.ErrorIsNil)

	adminBytes, err := db.GetByKey(s.d, admin.Admins, []byte(key))
	if err != nil {
		return err
	}
	c.Assert(len(adminBytes), gc.Not(gc.Equals), 0)

	var tmpAdmin admin.Admin
	if err = json.Unmarshal(adminBytes, &tmpAdmin); err != nil {
		return err
	}

	c.Check(tmpAdmin.Email, gc.Equals, adm.Email)
	return nil
}

func (s *AdminSuite) TestGetByEmail(c *gc.C) {
	s.createAdmins(c)
	defer s.deleteAdmins(c)

	for i, t := range []struct {
		should      string
		admin       t.TestAdmin
		expectError string
	}{{
		should: "fetch an admin by email",
		admin:  s.admins["bob"],
	}, {
		should:      "not fetch an admin that doesn't exist",
		admin:       t.TestAdmin{Email: "123456"},
		expectError: `admin for email 123456: user not found`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		if t.expectError == "" {
			c.Check(s.getByEmailTests(t.admin, c), jc.ErrorIsNil)
		} else {
			c.Check(s.getByEmailTests(t.admin, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *AdminSuite) getByEmailTests(adm t.TestAdmin, c *gc.C) error {
	tmpAdmin, err := admin.GetByEmail(s.d, adm.Email)
	if err != nil {
		return err
	}

	c.Check(tmpAdmin.Key, gc.Equals, adm.Key)
	return nil
}

func (s *AdminSuite) TestIsAdmin(c *gc.C) {
	s.createAdmins(c)
	defer s.deleteAdmins(c)

	for i, t := range []struct {
		should      string
		admin       t.TestAdmin
		expectError string
	}{{
		should: "validate an existing admin",
		admin:  s.admins["bob"],
	}, {
		should:      "not validate a nonexistent admin",
		admin:       t.TestAdmin{Key: "123456"},
		expectError: `admin for key 123456: user not found`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		if t.expectError == "" {
			c.Check(admin.IsAdmin(s.d, t.admin.Key), jc.ErrorIsNil)
		} else {
			c.Check(admin.IsAdmin(s.d, t.admin.Key), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *AdminSuite) TestIsAdminEmail(c *gc.C) {
	s.createAdmins(c)
	defer s.deleteAdmins(c)

	for i, t := range []struct {
		should      string
		admin       t.TestAdmin
		expectError string
	}{{
		should: "validate an existing admin",
		admin:  s.admins["bob"],
	}, {
		should:      "not validate a nonexistent admin",
		admin:       t.TestAdmin{Email: "foo@bar.com"},
		expectError: `admin for email foo@bar.com: user not found`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		if t.expectError == "" {
			c.Check(admin.IsAdminEmail(s.d, t.admin.Email), jc.ErrorIsNil)
		} else {
			c.Check(admin.IsAdminEmail(s.d, t.admin.Email), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *AdminSuite) TestDelete(c *gc.C) {
	s.createAdmins(c)

	for i, t := range []struct {
		should      string
		admin       t.TestAdmin
		expectError string
	}{{
		should: "delete an admin",
		admin:  s.admins["bob"],
	}, {
		should:      "not delete a nonexistent admin",
		admin:       t.TestAdmin{Key: "123456"},
		expectError: `admin for key 123456: user not found`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		if t.expectError == "" {
			c.Check(s.deleteTests(t.admin, c), jc.ErrorIsNil)
		} else {
			c.Check(s.deleteTests(t.admin, c), gc.ErrorMatches, t.expectError)
		}
	}
}

func (s *AdminSuite) deleteTests(adm t.TestAdmin, c *gc.C) error {
	if err := admin.Delete(s.d, adm.Key); err != nil {
		c.Logf("failed to delete admin %q", adm.Email)
		return err
	}

	err := admin.IsAdmin(s.d, adm.Key)
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf("admin for key %s: user not found", adm.Key))
	return nil
}

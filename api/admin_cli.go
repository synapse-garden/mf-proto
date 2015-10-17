package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/synapse-garden/mf-proto/admin"
	"github.com/synapse-garden/mf-proto/cli"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/util"

	"github.com/juju/errors"
	"gopkg.in/readline.v1"
)

func AdminCLI(d db.DB) cli.Binding {
	return func(c *cli.CLI) error {
		if err := db.SetupBuckets(d, admin.Buckets()); err != nil {
			return err
		}

		return c.AddCommands(&cli.Command{
			Name:        "create",
			Description: "create a new admin",
			Aliases:     []string{"c", "admin", "new"},
			Fn:          cliCreate(c, d),
		}, &cli.Command{
			Name:        "delete",
			Description: "delete an admin by email",
			Aliases:     []string{"d", "kill"},
			Fn:          cliDelete(d),
		})
	}
}

func getAdminEmail(rl *readline.Instance, d db.DB) (string, error) {
	none := ""

	rl.SetPrompt("Enter email: ")
	defer rl.SetPrompt("> ")

	email, err := rl.Readline()
	if err != nil {
		return none, err
	}

	if err := admin.IsAdminEmail(d, email); err != nil {
		if errors.IsUserNotFound(err) {
			return email, nil
		}
		return none, err
	}

	return none, fmt.Errorf("admin %s already exists", email)
}

func cliCreate(c *cli.CLI, d db.DB) cli.CommandFunc {
	return func(args ...string) (cli.Response, error) {
		var (
			createArgs = make(map[string]string)
			email      string
			pwhash     string
			err        error
			none       = cli.Response("")
			rl         = c.Rl
		)

		if rl == nil {
			return none, errors.New("tried to create admin using nil readline")
		}

		for _, arg := range args {
			splt := strings.SplitN(arg, "=", 2)
			if len(splt) < 2 {
				return "", errors.New("create args must be of the form pw=foo email=bar")
			}
			createArgs[splt[0]] = splt[1]
		}

		email, ok := createArgs["email"]
		if !ok {
			email, err = getAdminEmail(rl, d)
			if err != nil {
				return none, err
			}
		}

		pw, ok := createArgs["pw"]
		if !ok {
			var b []byte
			b, err = rl.ReadPassword("Enter password: ")
			if err != nil {
				return none, err
			}
			pw = string(b)
		}

		pwhash = string(util.SaltedHash(string(pw), time.Now().String()))

		key, err := admin.Create(d, email, pwhash)
		if err != nil {
			return none, err
		}

		return cli.Response("key: " + string(key)), nil
	}
}

func cliDelete(d db.DB) cli.CommandFunc {
	return func(args ...string) (cli.Response, error) {
		if len(args) != 1 {
			return "", errors.New("delete takes an email as its arg")
		}

		if err := admin.DeleteByEmail(d, args[0]); err != nil {
			return "", err
		}

		return cli.Response(fmt.Sprintf("admin %s deleted ok", args[0])), nil
	}
}

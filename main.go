package main

import (
	"log"

	"github.com/boltdb/bolt"
	"github.com/synapse-garden/mf-proto/api"
	"github.com/synapse-garden/mf-proto/cli"
)

func main() {
	d, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatalf("setting up db failed: %s", err.Error())
	}
	defer d.Close()

	c, err := cli.NewCLI(
		api.AdminCLI(d),
	)

	runHTTPListeners(d)
	c.Admin()
}

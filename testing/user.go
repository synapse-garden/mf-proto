package testing

import "github.com/synapse-garden/mf-proto/util"

type TestUser struct {
	Email     string
	Pwhash    string
	LoginKey  string
	LoginHash string
}

type TestAdmin struct {
	Email  string
	Pwhash string
	Key    util.Key
}

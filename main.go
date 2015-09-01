package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/boltdb/bolt"
	htr "github.com/julienschmidt/httprouter"
	"github.com/synapse-garden/mf-proto/api"
	"github.com/synapse-garden/mf-proto/db"
)

type Command struct {
	command, args string
}

func main() {
	d, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatalf("setting up db failed: %s", err.Error())
	}
	defer d.Close()

	go runHTTPListeners(d)

	cliAdmin()
}

func cliAdmin() {
	waitC := make(chan struct{})

	inputCommands := make(chan Command)
	go readCommands(waitC, inputCommands)

	commands := map[string]*regexp.Regexp{"quit": regexp.MustCompile("quit|exit|bye")}

	for command := range inputCommands {
		args, command := command.args, command.command
		_ = args // go die, compiler
		switch {
		case commands["quit"].MatchString(command):
			log.Printf("exiting program")
			os.Exit(0)
		default:
			log.Printf("invalid command %s", command)
		}
		waitC <- struct{}{}
	}
}

func readCommands(waitC chan struct{}, c chan Command) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		command, err := reader.ReadString('\n')
		if err != nil {
			log.Panic(err)
		}

		parts := strings.Split(command, " ")
		command, args := parts[0], strings.Join(parts[1:], " ")
		commandStruct := Command{command: command, args: args}

		c <- commandStruct
		<-waitC
	}
}

func runHTTPListeners(d db.DB) {
	httpMux := htr.New()
	httpsMux := htr.New()

	// httpMux.SetupRoutes()
	err := api.SetupRoutes(
		httpsMux,
		api.Auth(d),
		api.User(d),
		api.Task(d),
	)

	if err != nil {
		log.Fatalf("router setup failed: %s\n", err.Error())
	}

	var (
		httpErr  = make(chan error)
		httpsErr = make(chan error)
	)

	println("Listening on HTTP 25000")
	println("Listening on HTTPS 25001")

	go func() { httpErr <- http.ListenAndServeTLS(":25001", "cert.pem", "key.key", httpsMux) }()
	go func() { httpsErr <- http.ListenAndServe(":25000", httpMux) }()

	var e error
	select {
	case e = <-httpErr:
	case e = <-httpsErr:
	}
	log.Fatal("error serving http(s): %s", e.Error())
}

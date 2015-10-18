package cli

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
)

// Command is a CLI command which runs a given function Fn when the command
// is issued.
type Command struct {
	Name, Description string
	Aliases           []string
	Fn                CommandFunc
}

// Response is the type that returns from a Command.
type Response string

// CommandFunc is the type which a Command's Fn must be.
type CommandFunc func(...string) (Response, error)

func DefaultCommands(c *CLI) error {
	return c.AddCommands(&Command{
		Name:        "quit",
		Aliases:     []string{"q", "exit", "bye"},
		Description: "Exit",
		Fn:          quit,
	}, &Command{
		Name:        "help",
		Aliases:     []string{"?", "h", "usage"},
		Description: "Show usage syntax",
		Fn:          help(c),
	})
}


func help(c *CLI) CommandFunc {
	return func(args ...string) (Response, error) {
		usage := bytes.NewBufferString("Usage:\n")

		for name, command := range c.Commands {
			aliases := strings.Join(command.Aliases, ", ")
			usage.WriteString(fmt.Sprintf("\t%s (%s) - %s\n", name, aliases, command.Description))
		}

		return Response(usage.String()), nil
	}
}

func quit(args ...string) (Response, error) {
	log.Print("Bye!")
	os.Exit(0)
	return "", nil
}

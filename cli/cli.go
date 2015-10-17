package cli

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/readline.v1"
)

// Binding is a function for setting up new Commands from other packages.
type Binding func(*CLI) error

// CLI is a set of Commands and their aliases.
type CLI struct {
	Commands map[string]*Command

	aliases map[string]*Command

	Rl *readline.Instance
}

func NewCLI(bindings ...Binding) (*CLI, error) {
	c := &CLI{
		Commands: make(map[string]*Command),
		aliases:  make(map[string]*Command),
	}

	innerBindings := append([]Binding{DefaultCommands}, bindings...)
	for _, b := range innerBindings {
		if err := b(c); err != nil {
			return nil, err
		}
	}

	items := make([]*readline.PrefixCompleter, 0)
	for name := range c.Commands {
		items = append(items, readline.PcItem(name))
	}

	completer := readline.NewPrefixCompleter(items...)

	var err error
	c.Rl, err = readline.NewEx(&readline.Config{
		Prompt:       "> ",
		HistoryFile:  "/tmp/readline.tmp",
		AutoComplete: completer,
	})
	if err != nil {
		return nil, err
	}

	log.SetOutput(c.Rl.Stdout())

	return c, nil
}

func (c *CLI) AddCommands(cms ...*Command) error {
	for _, cm := range cms {
		if _, ok := c.Commands[cm.Name]; ok {
			return fmt.Errorf("command %q already defined", cm.Name)
		}

		c.Commands[cm.Name] = cm
		c.aliases[cm.Name] = cm
		for _, alias := range cm.Aliases {
			if _, ok := c.aliases[alias]; ok {
				return fmt.Errorf("alias %q already defined", alias)
			}

			c.aliases[alias] = cm
		}
	}

	return nil
}

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

// Admin reads commands from stdin and executes them.
func (c *CLI) Admin() {
	var (
		response Response
		action   CommandFunc
		waitC    = make(chan struct{})
		err      error
	)

	if c.Rl == nil {
		c.Rl, err = readline.NewEx(&readline.Config{
			Prompt:      "> ",
			HistoryFile: "/tmp/readline.tmp",
		})
	}

	inputCommands := make(chan string)
	go readCommands(waitC, inputCommands, c.Rl)

	for command := range inputCommands {
		if command == "" {
			waitC <- struct{}{}
			continue
		}

		parts := strings.Split(command, " ")
		command, args := parts[0], parts[1:]

		cm, ok := c.aliases[command]
		if !ok {
			log.Printf("unknown command %q", command)
			action = help(c)
		} else {
			action = cm.Fn
		}

		response, err = action(args...)
		if err != nil {
			log.Print(err)
		}
		log.Print(response)
		waitC <- struct{}{}
	}
}

func readCommands(waitC chan struct{}, c chan string, rl *readline.Instance) {
	for {
		command, err := rl.Readline()
		if err != nil {
			log.Panic(err)
		}
		command = strings.TrimSpace(command)

		c <- command
		<-waitC
	}
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

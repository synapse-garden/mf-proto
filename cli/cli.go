package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
)

//Command -
type Command struct {
	Name, Description string
	Aliases           []string
	Fn                commandFunc
}

//Response -
type Response string
type commandFunc func(...string) (Response, error)

var availableCommands = make(map[string]*Command)
var commandMap = make(map[string]*Command)

func makeAvailableCommands() {
	for name, command := range map[string]*Command{
		"quit": &Command{
			Name:        "quit",
			Aliases:     []string{"exit", "bye"},
			Description: "Exit",
			Fn:          quit,
		},
		"help": &Command{
			Name:        "help",
			Aliases:     []string{"?", "usage"},
			Description: "Show usage syntax",
			Fn:          help,
		},
	} {
		availableCommands[name] = command
	}
}

func makeCommandMap() {
	for name, command := range availableCommands {
		commandMap[name] = command
		for _, alias := range command.Aliases {
			commandMap[alias] = command
		}
	}
}

func init() {
	makeAvailableCommands()
	makeCommandMap()
}

//MatchCommand - match a command or an alias to the actual command struct
func MatchCommand(search string) (*Command, error) {
	if command, ok := commandMap[search]; ok {
		return command, nil
	}

	return nil, fmt.Errorf("Command %q not found", search)
}

//Admin - read commands from stdin and execute them
func Admin() {
	waitC := make(chan struct{})

	inputCommands := make(chan string)
	go readCommands(waitC, inputCommands)

	for command := range inputCommands {
		parts := strings.Split(command, " ")
		command, args := parts[0], parts[1:]

		action, err := MatchCommand(command)
		if err != nil {
			log.Printf("error: %s", err.Error())
			action, _ = MatchCommand("help")
		}

		response, err := action.Fn(args...)
		if err != nil {
			log.Print(err)
		}
		log.Print(response)
		waitC <- struct{}{}
	}
}

func readCommands(waitC chan struct{}, c chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		command, err := reader.ReadString('\n')
		if err != nil {
			log.Panic(err)
		}
		command = strings.TrimSpace(command)

		c <- command
		<-waitC
	}
}

func help(args ...string) (Response, error) {
	usage := bytes.NewBufferString("Usage:\n")

	for name, command := range availableCommands {
		usage.WriteString(fmt.Sprintf("\t%s - %s\n", name, command.Description))
	}

	return Response(usage.String()), nil
}

func quit(args ...string) (Response, error) {
	log.Print("Bye!")
	os.Exit(0)
	return "", nil
}

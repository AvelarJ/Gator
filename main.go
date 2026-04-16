package main

import (
	"fmt"
	"os"

	"example.com/gator/internal/config"
)

type state struct {
	*config.Config
}

type command struct {
	Name string
	Args []string
}

// Struct holds all usable commands
type commands struct {
	Handlers map[string]func(*state, command) error
}

// Run executes the command handler for the given command
func (c *commands) run(s *state, cmd command) error {
	if handler, ok := c.Handlers[cmd.Name]; ok {
		return handler(s, cmd)
	}
	return fmt.Errorf("unknown command: %s", cmd.Name)
}

// Register adds a new command handler to the commands struct
func (c *commands) register(name string, f func(*state, command) error) {
	c.Handlers[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("username required")
	}

	username := cmd.Args[0]
	s.Config.SetUser(username)

	fmt.Println("Logged in as", username)
	return nil
}

// Main function loop
func main() {
	conf, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
	}

	conf_state := state{Config: &conf}

	// Create a fresh commands struct
	cmds := commands{Handlers: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)

	// Parse command line arguments
	args := os.Args
	if len(args) < 2 { // 2 args minimum: gator <command> [args]
		fmt.Println("Usage: gator <command> [args]")
		os.Exit(1)
	}
	cmd := command{Name: args[1], Args: args[2:]}
	// Run the command and return any errors
	if err := cmds.run(&conf_state, cmd); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	//fmt.Println(conf_state.Config)

}

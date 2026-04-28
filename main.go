package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"

	"database/sql"

	"github.com/AvelarJ/Gator/internal/database"

	"github.com/AvelarJ/Gator/internal/config"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
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

// Handler functions for each command

// Register a new user if not already in db
func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("username required")
	}

	ctx := context.Background()

	uuid := uuid.New()
	username := sql.NullString{String: cmd.Args[0], Valid: true}
	//Check if username already exists
	oldUser, err := s.db.GetUser(ctx, username)
	if err == nil && oldUser.Name == username {
		fmt.Println("user already exists")
		os.Exit(1)
	}

	currUser, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid,
		Name:      username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	s.cfg.SetUser(currUser.Name.String)
	fmt.Println("Registered as", currUser.Name.String)
	fmt.Println(currUser)
	return nil
}

// Login an existing user
func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("username required")
	}

	username := sql.NullString{String: cmd.Args[0], Valid: true}

	ctx := context.Background()

	_, err := s.db.GetUser(ctx, username)
	if err != nil {
		fmt.Println("user does not exist")
		os.Exit(1)
	}

	s.cfg.SetUser(username.String)

	fmt.Println("Logged in as", username.String)
	return nil
}

// Main function loop
func main() {
	conf, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
	}

	//conf.Database_url = "postgres://jordanavelar:@localhost:5432/gator"

	// Opening a postgres database connection from the config
	db, err := sql.Open("postgres", conf.Database_url)
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}
	defer db.Close()

	// New database queries struct
	dbQueries := database.New(db)

	conf_state := state{db: dbQueries, cfg: &conf}

	// Create a fresh commands struct
	cmds := commands{Handlers: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)

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

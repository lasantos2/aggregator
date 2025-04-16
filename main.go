package main
import _ "github.com/lib/pq"
import (
	"fmt"
	"log"
	"context"
	"errors"
	"time"
	"os"
	"database/sql"
	"github.com/lasantos2/aggregator/internal/database"
	"github.com/lasantos2/aggregator/internal/config"
	"github.com/google/uuid"
)

type state struct {
	db *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commMap map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error){
	_, ok := c.commMap[name]
	if ok { // command already exists
		return
	}

	c.commMap[name] = f
}

func (c *commands) run(s *state, cmd command) error {

	err := c.commMap[cmd.name](s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func handleReset(s *state, cmd command) error {
	
	err := s.db.Reset(context.Background())
	if err != nil {
		os.Exit(1)
		return err
	}

	return nil
}

func handleLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("Username Required")
	}

	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		os.Exit(1)
		return err
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		os.Exit(1)
		return err
	}

	fmt.Println("User has been set!")
	return nil
}

func handleRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("Username Required")
	}

	Newuser := database.CreateUserParams{uuid.New(), time.Now(),time.Now(), cmd.args[0]}

	_, err := s.db.CreateUser(context.Background(),Newuser)

	if err != nil {
		fmt.Println("User already exists")
		return err
	}
	err = s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	//fmt.Println("User has been set!")
	return nil
}

func main() {
	cfg, err := config.Read()
	dbURL := cfg.DBURL

	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	stateInst := state{dbQueries, &cfg}
	commandsInst := commands{make(map[string]func(*state, command) error)}

	commandsInst.register("login", handleLogin)
	commandsInst.register("register", handleRegister)
	commandsInst.register("reset", handleReset)

	args := os.Args

	if len(args) < 2 {
		fmt.Println("not enough arguments provided! Ex : gator {command} {args}")
		os.Exit(1)
	}

	commandName := args[1]
	commandArgs := args[2:]

	commandInst := command{commandName, commandArgs}

	err = commandsInst.run(&stateInst, commandInst)
	if err != nil{
		fmt.Println(err)
		os.Exit(1)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	//fmt.Printf("Read config again: %+v\n", cfg)
}
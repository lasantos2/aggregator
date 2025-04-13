package main

import (
	"fmt"
	"log"
	
	"github.com/lasantos2/aggregator/internal/config"
)

type state struct {
	cfg *Config
}

type command struct {
	name string
	args []string
}

func handleLogin(s *state, cmd command) error {
	if len(args) == 0 {
		return error.NewError("Username Required")
	}

	err := s.cfg.SetUser(args)
	if err != nil {
		return err
	}

	fmt.Println("User has been set!")
}

func main() {
	cfg, err := config.Read()

	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	fmt.Printf("Read config: %+v\n", cfg)

	err = cfg.SetUser("Luis")

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	fmt.Printf("Read config again: %+v\n", cfg)
}
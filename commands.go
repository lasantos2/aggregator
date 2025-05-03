package main

import "errors"

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
	function, ok := c.commMap[cmd.name]
	if !ok {
		return errors.New("Command not found")
	}

	return function(s, cmd)
}
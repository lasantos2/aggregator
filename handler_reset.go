package main

import (
	"context"
	"fmt"
)

func handleReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		return fmt.Errorf("Coudln't delete user")
	}
	fmt.Println("Database reset successful")
	return nil
}

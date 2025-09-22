package main

import (
	"blog/internal/config"
	"blog/internal/database"
	"context"
	"errors"
	"fmt"
)

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		if s.Config == nil {
			cfg, err := config.Read()
			if err != nil {
				return err
			}
			s.Config = &cfg
		}

		if s.Config.User == "" {
			return errors.New("no user is currently logged in; please log in first")
		}

		ctx := context.Background()
		user, err := s.db.GetUser(ctx, s.Config.User)
		if err != nil {
			return fmt.Errorf("failed to retrieve current user: %v", err)
		}

		return handler(s, cmd, user)
	}
}

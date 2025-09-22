package main

import (
	"blog/internal/config"
	"blog/internal/database"
)

type State struct {
	db     *database.Queries
	Config *config.Config
}

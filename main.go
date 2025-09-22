package main

import (
	"blog/internal/config"
	"blog/internal/database"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	var newState State
	newState.Config = &cfg

	db, err := sql.Open("postgres", newState.Config.URL)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		os.Exit(1)
		return
	}
	defer db.Close()

	dbQueries := database.New(db)
	newState.db = dbQueries

	var cmds Commands
	cmds.commands = make(map[string]func(*State, Command) error)
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAggregate)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerListFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollowFeed))
	cmds.register("following", middlewareLoggedIn(handlerListFollows))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))
	cmds.register("browse", middlewareLoggedIn(handlerBrowsePosts))

	if len(os.Args) < 2 {
		fmt.Println("No command provided")
		os.Exit(1)
		return
	}

	enteredCommand := os.Args[1]
	args := []string{}

	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	if err := cmds.run(&newState, Command{name: enteredCommand, args: args}); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
		return
	}
}

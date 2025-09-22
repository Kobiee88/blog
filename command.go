package main

import (
	"blog/internal/config"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"blog/internal/database" // Add this import for the database package

	"github.com/google/uuid" // Add this import for uuid.New()
)

type Command struct {
	name string
	args []string
}

type Commands struct {
	commands map[string]func(*State, Command) error
}

func (c *Commands) register(name string, handler func(*State, Command) error) {
	if c.commands == nil {
		c.commands = make(map[string]func(*State, Command) error)
	}
	c.commands[name] = handler
}

func (c *Commands) run(s *State, cmd Command) error {
	if handler, exists := c.commands[cmd.name]; exists {
		return handler(s, cmd)
	}
	return fmt.Errorf("unknown command: %s", cmd.name)
}

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.args) < 1 {
		return errors.New("username argument is required")
	}

	username := cmd.args[0]

	if s.Config == nil {
		cfg, err := config.Read()
		if err != nil {
			return err
		}
		s.Config = &cfg
	}

	ctx := context.Background()

	_, err := s.db.GetUser(ctx, username)
	if err != nil {
		return err
	}

	if err := s.Config.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("User set to: %s\n", username)

	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.args) < 1 {
		return errors.New("username argument is required")
	}

	username := cmd.args[0]

	if s.Config == nil {
		cfg, err := config.Read()
		if err != nil {
			return err
		}
		s.Config = &cfg
	}

	ctx := context.Background()

	_, err := s.db.GetUser(ctx, username)
	if err == nil {
		return fmt.Errorf("user %s already exists", username)
	}

	newUserParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	}

	_, err = s.db.CreateUser(ctx, newUserParams)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	if err := s.Config.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("User %s registered successfully\n", username)
	return nil
}

func handlerReset(s *State, cmd Command) error {
	ctx := context.Background()

	if err := s.db.DeleteUsers(ctx); err != nil {
		return fmt.Errorf("failed to delete users: %v", err)
	}

	//ctx = context.Background()
	if err2 := s.db.DeleteFeeds(ctx); err2 != nil {
		return fmt.Errorf("failed to delete feeds: %v", err2)
	}

	if err3 := s.db.DeleteFeedFollows(ctx); err3 != nil {
		return fmt.Errorf("failed to delete feed follows: %v", err3)
	}

	fmt.Println("All users, feeds and follows have been deleted.")
	return nil
}

func handlerUsers(s *State, cmd Command) error {
	ctx := context.Background()

	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve users: %v", err)
	}

	currentUser := s.Config.User

	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	fmt.Println("Registered Users:")
	for _, user := range users {
		if user.Name == currentUser {
			fmt.Printf("- %s (current)\n", user.Name)
			continue
		}
		fmt.Printf("- %s\n", user.Name)
	}

	return nil
}

func handlerAggregate(s *State, cmd Command) error {
	time_between_reqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}

	ticker := time.NewTicker(time_between_reqs)
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func handlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("feed name and URL arguments are required")
	}

	feedName := cmd.args[0]
	feedURL := cmd.args[1]

	newFeedParams := database.CreateFeedParams{
		ID:            uuid.New(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		LastFetchedAt: sql.NullTime{},
		Name:          feedName,
		Url:           feedURL,
		UserID:        user.ID,
	}

	_, err := s.db.CreateFeed(context.Background(), newFeedParams)
	if err != nil {
		return fmt.Errorf("failed to create feed: %v", err)
	}

	fmt.Printf("Feed '%s' (URL: '%s') added successfully for user '%s'\n", feedName, feedURL, user.Name)

	handlerFollowFeed(s, Command{name: "follow", args: []string{feedURL}}, user)

	return nil
}

func handlerListFeeds(s *State, cmd Command) error {
	ctx := context.Background()

	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve feeds: %v", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}

	fmt.Println("Feeds:")
	for _, f := range feeds {
		ctx = context.Background()
		feedUser, err := s.db.GetUserById(ctx, f.UserID)
		if err != nil {
			return fmt.Errorf("failed to retrieve user: %v", err)
		}
		fmt.Printf("- ID: %d Name: %s URL: %s User: %s\n", f.ID, f.Name, f.Url, feedUser.Name)
	}

	return nil
}

func handlerFollowFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("feed URL argument is required")
	}
	feedURL := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("failed to retrieve feed by URL: %v", err)
	}

	newFeedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), newFeedFollowParams)
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %v", err)
	}

	fmt.Printf("User '%s' is now following feed '%s' (URL: '%s')\n", user.Name, feed.Name, feed.Url)
	return nil
}

func handlerListFollows(s *State, cmd Command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve feed follows: %v", err)
	}

	if len(follows) == 0 {
		fmt.Printf("User '%s' is not following any feeds.\n", s.Config.User)
		return nil
	}

	fmt.Printf("Feeds followed by user '%s':\n", user.Name)
	for _, follow := range follows {
		fmt.Printf("- %s\n", follow.FeedName)
	}

	return nil
}

func handlerUnfollowFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("feed URL argument is required")
	}
	feedURL := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("failed to retrieve feed by URL: %v", err)
	}

	newDeleteFeedFollowParams := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.db.DeleteFeedFollow(context.Background(), newDeleteFeedFollowParams)
	if err != nil {
		return fmt.Errorf("failed to unfollow feed: %v", err)
	}

	fmt.Printf("User '%s' has unfollowed feed '%s' (URL: '%s')\n", user.Name, feed.Name, feed.Url)
	return nil
}

func handlerBrowsePosts(s *State, cmd Command, user database.User) error {
	limit := 2
	if len(cmd.args) >= 1 {
		limit, _ = strconv.Atoi(cmd.args[0])
	}

	newGetPostsParams := database.GetPostsForUserParams{
		ID:    user.ID,
		Limit: int32(limit),
	}

	posts, err := s.db.GetPostsForUser(context.Background(), newGetPostsParams)
	if err != nil {
		return fmt.Errorf("failed to retrieve posts: %v", err)
	}

	if len(posts) == 0 {
		fmt.Printf("No posts found for user '%s'.\n", user.Name)
		return nil
	}

	fmt.Printf("Posts for user '%s':\n", user.Name)
	for _, post := range posts {
		fmt.Printf("- %s (%s)\n", post.Title, post.PublishedAt.Format(time.RFC1123))
	}

	return nil
}

func scrapeFeeds(s *State) error {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get next feed to fetch: %v", err)
	}

	feed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch RSS feed: %v", err)
	}

	newFetchedParams := database.MarkFeedAsFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt:     time.Now(),
		ID:            nextFeed.ID,
	}

	err = s.db.MarkFeedAsFetched(context.Background(), newFetchedParams)
	if err != nil {
		return fmt.Errorf("failed to mark feed as fetched: %v", err)
	}

	fmt.Printf("Fetched Feed Title: %s\n", feed.Channel.Title)

	for _, item := range feed.Channel.Item {
		if _, err := s.db.GetPostByUrl(context.Background(), item.Link); err == nil {
			// Post already exists, skip it
			continue
		}
		fmt.Printf("- %s (%s)\n", item.Title, item.PubDate)
		// Parse the publication date string to time.Time
		pubTime, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			pubTime, err = time.Parse(time.RFC1123, item.PubDate)
			if err != nil {
				pubTime, err = time.Parse(time.RFC822Z, item.PubDate)
				if err != nil {
					pubTime, err = time.Parse(time.RFC822, item.PubDate)
					if err != nil {
						return fmt.Errorf("failed to parse publication date: %v", err)
					}
				}
			}
		}
		newPostParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			FeedID:      nextFeed.ID,
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: pubTime,
		}

		_, err = s.db.CreatePost(context.Background(), newPostParams)
		if err != nil {
			return fmt.Errorf("failed to create post: %v", err)
		}
	}

	return nil
}

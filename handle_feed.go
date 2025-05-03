package main

import (
	"fmt"
	"time"
	""
)


func handleFeedGet(s *state, cmd command) error {

	if len(cmd.args) != 1 {
		return errors.New("Need time input: 1s, 1m, 1h")
	}

	timeBetweenReqs, err:= time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(timeBetweenReqs)

	for ;; <- ticker.C {
		scrapeFeeds(s)
	}

	return nil
}

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.name)
	}

	feedname := cmd.args[0]
	feedurl := cmd.args[1]

	// connect feed to that user
	NewFeed := database.CreateFeedParams{uuid.New(), time.Now(),time.Now(),feedname,feedurl,user.ID}
	feed, err := s.db.CreateFeed(context.Background(),NewFeed)
	if err != nil {
		return fmt.Errof("couldn't create feed: %w", err)
	}

	// Create feedFollow record for current user
	newFeedFollow := database.CreateFeedFollowParams{
		uuid.New(),
		time.Now(),
		time.Now(),
		feed.ID,
		user.ID}
	
	_, err = s.db.CreateFeedFollow(context.Background(), newFeedFollow)
	if err != nil {
		return fmt.Errof("couldn't create feed follow: %w", err)
	}

	return nil
}

func handleShowFeeds(s *state, cmd command) error {

	Feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errof("couldn't get feed: %w", err)
	}

	for _, feed := range Feeds {
		fmt.Println(feed.Name)
		username, err := s.db.GetFeedUser(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Println(username)
	}
	return nil
}
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
	"io"
	"net/http"
	"encoding/xml"
	"github.com/lasantos2/aggregator/internal/database"
	"github.com/lasantos2/aggregator/internal/config"
	"github.com/google/uuid"
	"html"
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

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error ) func(*state, command) error {
	// get current user from database

	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
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

	return nil
}

func handleUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n",user.Name)
		} else {
			fmt.Printf("* %s\n",user.Name)
		}
	}

	return nil
}

func handleFeedGet(s *state, cmd command) error {
	feedUrlString := "https://www.wagslane.dev/index.xml"
	rssFeed, err := fetchFeed(context.Background(), feedUrlString)
	if err != nil {
		return err
	}

	fmt.Println(rssFeed)

	return nil
}

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		os.Exit(1)
		return errors.New("Arguments are [name] [url]")
	}

	feedname := cmd.args[0]
	feedurl := cmd.args[1]

	// connect feed to that user
	NewFeed := database.CreateFeedParams{uuid.New(), time.Now(),time.Now(),feedname,feedurl,user.ID}
	feed, err := s.db.CreateFeed(context.Background(),NewFeed)
	if err != nil {
		os.Exit(1)
		return errors.New("Feed couldn't be created")
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
		fmt.Println(err)
		os.Exit(1)
		return errors.New("Feed Follow Record not created")
	}
	

	return nil
}

func handleShowFeeds(s *state, cmd command) error {

	Feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		os.Exit(1)
		return err
	}

	for _, feed := range Feeds {
		fmt.Println(feed.Name)
		username, err := s.db.GetFeedUser(context.Background(), feed.UserID)
		if err != nil {
			os.Exit(1)
			return err
		}
		fmt.Println(username)
	}
	return nil
}

func handleFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		os.Exit(1)
		return errors.New("Arguments are command [url]")
	}

	url := cmd.args[0]

	feeds,err  := s.db.GetFeeds(context.Background())
	if err != nil {
		os.Exit(1)
		return errors.New("Feed not found")
	}

	foundFeed := database.Feed{}

	for _, feed := range feeds {
		if feed.Url == url {
			foundFeed = feed
			break;
		}
	}
	
	newFeedFollow := database.CreateFeedFollowParams{
	uuid.New(),
	time.Now(),
	time.Now(),
	foundFeed.ID,
	user.ID}

	feedsFollowed, err := s.db.CreateFeedFollow(context.Background(), newFeedFollow)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return errors.New("No feed to follow")
	}

	for _, r := range feedsFollowed {
		fmt.Println(r.FeedName)
		fmt.Println(r.UserName)
	}

	return nil
}

func handleShowFollowing(s *state, cmd command, user database.User) error {
	followedFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Println(user.Name)
	for _, feed := range followedFeeds {
		fmt.Println(feed.FeedName)
		
	}

	return nil
}

func handleUnfollow(s *state, cmd command, user database.User) error {

	if len(cmd.args) != 1 {
		return errors.New("Need URL to unfollow")
	}

	url := cmd.args[0]

	deleteParams := database.DeleteFeedParams{}
	deleteParams.ID = user.ID
	deleteParams.Url = url


	err := s.db.DeleteFeed(context.Background(), deleteParams)
	if err != nil {
		log.Fatalf("feed doesn't exist, or username not valid?")
		return err
	}

	fmt.Println("Feed successfully unfollowed for ", user.Name)
	return nil
}

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {
	client := &http.Client{}

	resp, err := client.Get(feedUrl)
	if err != nil {
	
		return &RSSFeed{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)
	
	if err != nil {
	
		return &RSSFeed{}, err
	}

	req.Header.Set("User-Agent","gator")
	resp, err = client.Do(req)

	if err != nil {
		return &RSSFeed{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	rssFeed := RSSFeed{}

	xml.Unmarshal(body, &rssFeed)
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

	for ind, _ := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[ind].Title = html.UnescapeString(rssFeed.Channel.Item[ind].Title)
		rssFeed.Channel.Item[ind].Description = html.UnescapeString(rssFeed.Channel.Item[ind].Description)
	}

	return &rssFeed, nil

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
	commandsInst.register("users", handleUsers)
	commandsInst.register("agg", handleFeedGet)
	commandsInst.register("addfeed", middlewareLoggedIn(handleAddFeed))
	commandsInst.register("feeds", handleShowFeeds)
	commandsInst.register("follow", middlewareLoggedIn(handleFollowFeed))
	commandsInst.register("following", middlewareLoggedIn(handleShowFollowing))
	commandsInst.register("unfollow", middlewareLoggedIn(handleUnfollow))


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

}
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/lib/pq"

	"database/sql"

	"strconv"

	"github.com/AvelarJ/Gator/internal/database"

	"github.com/AvelarJ/Gator/internal/config"

	"github.com/AvelarJ/Gator/internal/rss"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	Name string
	Args []string
}

// Struct holds all usable commands
type commands struct {
	Handlers map[string]func(*state, command) error
}

// Run executes the command handler for the given command
func (c *commands) run(s *state, cmd command) error {
	if handler, ok := c.Handlers[cmd.Name]; ok {
		return handler(s, cmd)
	}
	return fmt.Errorf("unknown command: %s", cmd.Name)
}

// Register adds a new command handler to the commands struct
func (c *commands) register(name string, f func(*state, command) error) {
	c.Handlers[name] = f
}

// Agregation function to fetch the next feed in an interval
func scrapeFeeds(s *state) error {
	ctx := context.Background()

	// Find the next feed to fetch
	next, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("error fetching next feed:", err)
	}

	// Now mark the feed as fetched
	err = s.db.MarkFeedFetched(ctx, next.ID)
	if err != nil {
		return fmt.Errorf("error marking feed as fetched: %w", err)
	}

	// Now actually fetch the feed via url
	feed, err := rss.FetchFeed(ctx, next.Url)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}

	// Iterate over the feed and print the titles
	for _, item := range feed.Channel.Item {
		//fmt.Println(item.Title)
		// Need to fix publishedAt as it could be dif formats
		publishedAt, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			return fmt.Errorf("error parsing publishedAt: %w", err)
		}

		// Need to check if url is unique and if so ignore it

		_, err = s.db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: sql.NullTime{Time: publishedAt, Valid: true},
			FeedID:      next.ID,
		})
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				continue
			}
			return fmt.Errorf("error creating feed item: %w", err)
		}
	}

	return nil

}

// Handler functions for each command

// REMINDER TO REMOVE - TESTING PURPOSES ONLY
func handlerReset(s *state, _ command) error {
	ctx := context.Background()
	err := s.db.Reset(ctx)
	if err != nil {
		return fmt.Errorf("error resetting database: %w", err)
	}
	return nil
}

// Register a new user if not already in db
func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("username required")
	}

	ctx := context.Background()

	uuid := uuid.New()
	username := cmd.Args[0]
	//Check if username already exists
	oldUser, err := s.db.GetUser(ctx, username)
	if err == nil && oldUser.Name == username {
		fmt.Println("user already exists")
		os.Exit(1)
	}

	currUser, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid,
		Name:      username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	s.cfg.SetUser(currUser.Name)
	fmt.Println("Registered as", currUser.Name)
	fmt.Println(currUser)
	return nil
}

// Login an existing user
func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("username required")
	}

	username := cmd.Args[0]

	ctx := context.Background()

	_, err := s.db.GetUser(ctx, username)
	if err != nil {
		fmt.Println("user does not exist")
		os.Exit(1)
	}

	s.cfg.SetUser(username)

	fmt.Println("Logged in as", username)
	return nil
}

// Return a list of all users
func handlerGetUsers(s *state, _ command) error {
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("Unable to run GetUsers command")
	}
	for _, user := range users {
		if user.Name == s.cfg.Current_user {
			fmt.Println("*", user.Name, "(current)")
		} else {
			fmt.Println("*", user.Name)
		}
	}
	return nil
}

// In future will be used for aggregatting Multiple RSS feeds
// Now only uses constant URL
func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage: agg <interval> (eg. agg 10s)")
	}
	//ctx := context.Background()
	interval, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Invalid interval: %w", err)
	}

	fmt.Printf("Collecting feeds every %s\n", interval)
	// Create a ticker that will trigger scrapeFeeds over the given interval
	ticker := time.NewTicker(interval)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			fmt.Println("Error scraping feeds:", err)
			return err
		}
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	// At top get the current user
	ctx := context.Background()

	// Check args
	if len(cmd.Args) != 2 {
		return fmt.Errorf("Usage: addfeed <name> <url>")
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	// Create the feed in the database
	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("Unable to create feed\n", err)
	}

	// Now add the feed to the user's feed_follows table
	_, err = s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feed.ID,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("Unable to follow feed\n", err)
	}

	fmt.Println(feed)

	return nil
}

func handlerGetFeeds(s *state, _ command) error {
	ctx := context.Background()

	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get feeds", err)
	}
	for _, feed := range feeds {
		fmt.Println(feed)
	}
	return nil
}

// Command to follow a feed (Inserts into feed_follows table)
// Takes a url for the feed to be added
func handlerFeedFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Feed url required")
	}

	// params
	url := cmd.Args[0]
	ctx := context.Background()

	// Need to find a feed by url to get the feed_id
	feed, err := s.db.GetFeedUrl(ctx, url)
	if err != nil {
		return fmt.Errorf("Error retrieving feed\n", err)
	}

	// Main query to insert feed_follows record
	results, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed follow\n", err)
	}

	fmt.Println(results.FeedName, results.UserName)
	return nil

}

func handlerFollowingForUser(s *state, _ command, user database.User) error {

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Error retrieving feed follows\n", err)
	}

	for _, follow := range follows {
		fmt.Println(follow.FeedName, follow.UserName)
	}
	return nil
}

func handlerFeedUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage: feed unfollow <feed_url>")
	}

	feedURL := cmd.Args[0]
	feed, err := s.db.GetFeedUrl(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("Error retrieving feed: %w", err)
	}

	err = s.db.FeedUnfollow(context.Background(), database.FeedUnfollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error unfollowing feed: %w", err)
	}

	fmt.Printf("Unfollowed %s successfully\n", feedURL)
	return nil

}

func handlerBrowse(s *state, cmd command, user database.User) error {
	// Limit is optional, default to 2 if not provided
	var limit int
	if len(cmd.Args) != 1 {
		limit = 2
	} else { // Limit provided, convert to int
		var err error
		limit, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Invalid limit: %w", err)
		}
	}

	ctx := context.Background()
	// Get posts from database
	posts, err := s.db.GetPostUser(ctx, int32(limit))
	if err != nil {
		return fmt.Errorf("Error retrieving posts: %w", err)
	}
	// Print posts
	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("URL: %s\n", post.Url)
		fmt.Printf("Description: %s\n", post.Description.String)
		fmt.Printf("PublishedAt: %s\n", post.PublishedAt.Time.String())
		fmt.Println()
	}
	return nil
}

// MIDDLEWARE
// Middleware to add a logged in check higher function to the handlers that need it
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.Current_user) // Simple login check
		if err != nil {
			return fmt.Errorf("Error retrieving user info\n", err)
		}
		return handler(s, cmd, user)
	}
}

// Main function loop
func main() {
	conf, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
	}

	//conf.Database_url = "postgres://jordanavelar:@localhost:5432/gator"

	// Opening a postgres database connection from the config
	db, err := sql.Open("postgres", conf.Database_url)
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}
	defer db.Close()

	// New database queries struct
	dbQueries := database.New(db)

	conf_state := state{db: dbQueries, cfg: &conf}

	// Create a fresh commands struct
	cmds := commands{Handlers: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerGetFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFeedFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowingForUser))
	cmds.register("unfollow", middlewareLoggedIn(handlerFeedUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	// Parse command line arguments
	args := os.Args
	if len(args) < 2 { // 2 args minimum: gator <command> [args]
		fmt.Println("Usage: gator <command> [args]")
		os.Exit(1)
	}
	cmd := command{Name: args[1], Args: args[2:]}
	// Run the command and return any errors
	if err := cmds.run(&conf_state, cmd); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	//fmt.Println(conf_state.Config)

}

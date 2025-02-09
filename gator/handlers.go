package main

import(
	"github.com/lib/pq"
	"database/sql"
	"fmt"
	"errors"
	"context"
	"strings"
	"net/url"
	"github.com/fatih/color"
	"github.com/google/uuid"
)

func middlewareRequireLogin(handler func(s *state, cmd command, user uuid.UUID) error) func(*state, command) error {
    return func(s *state, cmd command) error {
    	thisUserID := s.config.CurrentUserID
    	if s.config.CurrentUserName == "" || thisUserID == (uuid.UUID{}){
			return errors.New("You must be logged in to use this command")
		}
        return handler(s, cmd, thisUserID)
    }
}

func handlerLogin(s *state, cmd command) error{
	if len(cmd.args) != 1{
		return fmt.Errorf("usage: %s", color.CyanString("login {username}"))
	}

	result, err := s.db.GetUser(context.Background(), cmd.args[0])
    if err != nil{
    	if errors.Is(err, sql.ErrNoRows) {
    		return fmt.Errorf("Not a user, use %s to add this user", color.YellowString("register " +cmd.args[0]))
    	}
        return fmt.Errorf("Unable to fetch user details from database: %w", err)
    }

	s.config.SetUser(cmd.args[0], result.ID)
	fmt.Printf("Username set to %s\n", color.YellowString(cmd.args[0]))
	return nil
}

func handlerRegister(s *state, cmd command) error{
	if len(cmd.args) != 1{
		return errors.New("usage: register {username}")
	}

	result, err := s.db.CreateUser(context.Background(), cmd.args[0])
    if err != nil{
    	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
    		return fmt.Errorf("A user with the name %s already exists",  color.YellowString(cmd.args[0]))
    	}
        return fmt.Errorf("Unable to create user: %w", err)
    }
    fmt.Printf("Registered %s (uuid: %s)\n", color.YellowString(result.Name), color.MagentaString(result.ID.String()))

	s.config.SetUser(cmd.args[0], result.ID)
	fmt.Printf("Username set to %s\n", color.YellowString(cmd.args[0]))
	return nil
}

func handlerReset(s *state, cmd command) error{
	if len(cmd.args) != 0{
		return errors.New("Unexpected arguments, usage: reset")
	}

	err := s.db.ResetUsers(context.Background())
    if err != nil{
        return fmt.Errorf("Unable to reset users table: %w", err)
    }

    s.config.SetUser("", uuid.UUID{})
    fmt.Println("Users reset!")
	return nil
}

func handlerListUsers(s *state, cmd command) error{
	if len(cmd.args) != 0{
		return errors.New("Unexpected arguments, usage: users")
	}

	results, err := s.db.GetUsers(context.Background())
    if err != nil{
        return fmt.Errorf("Unable to fetch users: %w", err)
    }

    for _,name := range results{
    	if strings.ToLower(name) == strings.ToLower(s.config.CurrentUserName){
    		fmt.Println(name + " (current)")
    		continue
    	}
    	fmt.Println(name)
    }

	return nil
}

func handlerAgg(s *state, cmd command) error{
	// if len(cmd.args) != 1{
	// 	return errors.New("usage: agg {url}")
	// }

	output, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil{
		return fmt.Errorf("Unable to fetch feed: %w", err)	
	}
	fmt.Print(output)
	return nil

}

func handlerAddFeed(s *state, cmd command, userID uuid.UUID) error{

	if len(cmd.args) != 2{
		return errors.New("usage: addfeed {feed_name} {feed_url}")
	}

	parsedUrl, err := url.Parse(cmd.args[1])
	if err != nil {
		return fmt.Errorf("Invalid feed url: %w.", err)
	}
	if !(parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" || parsedUrl.Scheme == "rss"){
		return fmt.Errorf("Invalid scheme '%s', must be http(s) or rss", parsedUrl.Scheme)
	}
	result, err := addFeed(context.Background(),s, cmd.args[0], cmd.args[1], userID)
	if err != nil{
		return fmt.Errorf("Unable to add feed: %w", err)	
	}

	_, err = addFeedFollow(s, userID,  result.ID)
	if err != nil{
		return fmt.Errorf("Added feed, but unable to follow: %w", err)	
	}

	fmt.Printf("%+v\n", result)

	return nil
}

func handlerListFeeds(s *state, cmd command) error{

	if len(cmd.args) != 0{
		return errors.New("Unexpected arguments, usage: feeds")
	}

	feeds, err := getFeeds(s)
	if err != nil {
		return fmt.Errorf("Unable to get feeds: %w.", err)
	}
	if len(feeds)==0{
		return errors.New("No feeds to list! Add one with addfeed {feed_name} {feed_url}")
	}
	for _,feed := range feeds{
		fmt.Printf("%s (%s) [User: %s]\n", feed.Name, feed.Url, feed.Username)
	}
	return nil
}

func handlerFollow(s *state, cmd command, userID uuid.UUID) error{

	if len(cmd.args) != 1{
		return errors.New("usage: follow {url}")
	}
	
	feed, err := getFeed(s, strings.ToLower(cmd.args[0]))
	if err != nil {
		return fmt.Errorf("Unable to get a feed for that url, use the add command?")
	}

    result, err := addFeedFollow(s, userID, feed.ID)
	if err != nil {
		return fmt.Errorf("Unable to follow feed: %w.", err)
	}

	fmt.Printf("You are now following %s, %s!", result.UserName, result.FeedName)
	return nil
}

func handlerListFeedFollow(s *state, cmd command, userID uuid.UUID) error{

	if len(cmd.args) != 0{
		return errors.New("Unexpected arguments, usage: following")
	}

	feeds, err := getUserFeedFollow(s, userID)
	if err != nil {
		return fmt.Errorf("Unable to get followed feeds: %w.", err)
	}
	if len(feeds)==0{
		return errors.New("No followed feeds to list!")
	}
	fmt.Println("You are following:")
	for _,feed := range feeds{
		fmt.Printf("%s (%s)\n", feed.FeedName, feed.Url)
	}
	return nil
}
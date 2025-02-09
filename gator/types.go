package main

import(
	"github.com/StupidWeasel/bootdev-blog-aggregator/gator/internal/config"
	"github.com/StupidWeasel/bootdev-blog-aggregator/gator/internal/database"
)

type state struct {
	config *config.Config
	db *database.Queries
}

type command struct {
    name string
    args []string
}

type commands struct{
	handlers map[string]func(*state, command) error 
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

type ListedFeed struct{
	Name	string
	URL		string
	User	string
}
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/StupidWeasel/bootdev-blog-aggregator/gator/internal/config"
	"github.com/StupidWeasel/bootdev-blog-aggregator/gator/internal/database"
	_ "github.com/lib/pq"
)
func main(){

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    goodbye := func(){
        fmt.Println("Goodbye!")
        cancel()
    }

    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
        <-sigChan
        goodbye()
    }()

    appState := state{
        config : &config.Config{
                    ConfigFile : ".gatorconfig.json",
                },
        context: ctx,
    }

    err := appState.config.ReadConfig()
    if err != nil{
        log.Fatalf("Unable to read config: %s", err)
    }


    if appState.config.PostBatchSize == 0 {
        appState.config.PostBatchSize = 100
    }


    db, err := sql.Open("postgres", appState.config.DbURL)
    if err != nil{
        log.Fatalf("Unable to connect to DB: %s", err)
    }
    appState.db = database.New(db)

    appCommands := commands{
        handlers: make(map[string]func(*state, command) error),
    }
    appCommands.register("login", handlerLogin)
    appCommands.register("register", handlerRegister)
    appCommands.register("reset", handlerReset)
    appCommands.register("users", handlerListUsers)
    appCommands.register("agg", handlerAgg)
    appCommands.register("addfeed", middlewareRequireLogin(handlerAddFeed))
    appCommands.register("feeds", handlerListFeeds)
    appCommands.register("follow", middlewareRequireLogin(handlerFollow))
    appCommands.register("following", middlewareRequireLogin(handlerListFeedFollow))
    appCommands.register("unfollow", middlewareRequireLogin(handlerRemoveFeedFollow))

    appCommands.register("browse", middlewareRequireLogin(handlerBrowseNext))
    appCommands.register("browse:next", middlewareRequireLogin(handlerBrowseNext))
    appCommands.register("next", middlewareRequireLogin(handlerBrowseNext))
    appCommands.register("browse:back", middlewareRequireLogin(handlerBrowseBack))
    appCommands.register("back", middlewareRequireLogin(handlerBrowseBack))

    theseArgs := os.Args[1:]

    if len(theseArgs)==0{
        fmt.Fprintf(os.Stderr, "Not enough arguments provided\n")
        os.Exit(1)
    }

    err = appCommands.run(&appState, command{
                                            name: theseArgs[0],
                                            args: theseArgs[1:],
                                        })
    if err != nil{
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

}

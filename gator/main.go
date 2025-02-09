package main

import(
	"database/sql"
	"fmt"
	"log"
	"os"
	_ "github.com/lib/pq"
	"github.com/StupidWeasel/bootdev-blog-aggregator/gator/internal/config"
	"github.com/StupidWeasel/bootdev-blog-aggregator/gator/internal/database"
)
func main(){

	appState := state{
		config : &config.Config{ 
					ConfigFile : ".gatorconfig.json",
				},
	}
	
	err := appState.config.ReadConfig()
	if err != nil{
		log.Fatalf("Unable to read config: %w", err)
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

# :bear: Boot.dev -- 17. Build a Blog Aggregator in Go
<!-- Hello. If you're reading this you're a beautiful and wonderful person. -->

![Tools with the Go Gopher](https://storage.googleapis.com/qvault-webapp-dynamic-assets/course_assets/fbuH9HC.png)

My work as part of the boot.dev Build a Blog Aggregator guided project.
[Check out boot.dev (with a referal from me)](https://wzl.to/boot.dev)  ***-or-*** [Check it out, without a referal](https://wzl.to/boot.dev_noref)

## :wrench: Requirements & Installation
#### Requirements
Requires [postgres (PostgreSQL) 17.2](https://www.postgresql.org/docs/current/tutorial-install.html) and [go version go1.23.5](https://go.dev/doc/install), older versions may work but are untested. Instructions written with Linux in mind, and tested on the wonderful [EndeavourOS](https://endeavouros.com/), but should work with any OS that meets the [go min specs](https://go.dev/wiki/MinimumRequirements).
#### Installation / building
 - Grab this repo either via [git clone](https://git-scm.com/docs/git-clone) (recommended) or via [automatically generated zip](https://github.com/StupidWeasel/bootdev-blog-aggregator/archive/refs/heads/main.zip).
 - Create a [new postgres database with psql](https://www.postgresql.org/docs/current/app-psql.html):
```
sudo -u postgres psql
CREATE DATABASE db_name_here;
\c db_name_here
ALTER USER desired_user PASSWORD 'a_nice_long_password_string';
```

 - Create a ~/.gatorconfig.json ***in your homedir,*** an example can be found in the root directory of the project, but it only needs to be the following to start with:

     {
      "config_file": ".gatorconfig.json",
      "db_url": "postgres://deired_user:a_nice_long_password_string@localhost:5432/db_name_here?sslmode=disable",
     }

 - Navigate to `root_of_project/gator/sql/` (ideally in the terminal) & run  `./goose.sh up` to setup the database, it should say something along the lines of `goose: no migrations to run. current version: 5` when done.
 - In `root_of_project/gator/` run either `go build` to create an executable within the directory or run `go install` to install it in your `$GOPATH/bin` directory. Either way `gator` will be created!
 
 There we go, all done!
## :computer: Commands
#### user commands
`gator register {username}` - Register a username

`gator login {username}` - Login as a registered user

`gator users` - List users

#### feed commands
`gator addfeed {feed_name} {feed_url}` - Add a feed (requires login) (automatically follows feed)

⠀⠀*Example:* `gator addfeed "Boot.dev Blog" "https://blog.boot.dev/index.xml"`

`gator feeds` - List All Feeds

`gator follow {feed_url}` - Follow/subscribe to a feed (requires login)

`gator unfollow {feed_url}` - Unfollow a feed (requires login)

`gator following` - Lists al feeds you are following

`gator agg {interval}` - Fetches new posts from the feed that has gone longest without an update. Runs until terminated.

⠀⠀*Example:* `gator agg 10m` - Fetches new posts every 10 minutes. Ctrl+C cleanely kills the process.

> [!CAUTION]
> Try not to hammer the servers you are fetching posts from, that's not very nice.

#### post/pagination commands
`gator next {optional limit between 1-10}` - View the next nth posts, defaults to 2. Sorted newest to oldest.

`gator back {optional limit between 1-10}` - View the previous nth posts, defaults to 2. Sorted newest to oldest.

⠀ 



*If for some odd reason you enjoyed reading this, you might [enjoy the courses on boot.dev even more](https://wzl.to/boot.dev_noref), you can even [tell them I sent you](https://wzl.to/boot.dev) if you want.*

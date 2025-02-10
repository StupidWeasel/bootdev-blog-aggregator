package main

import(
    "github.com/lib/pq"
    "database/sql"
    "fmt"
    "errors"
    "strings"
    "time"
    "net/url"
    "os"
    "strconv"
    "github.com/fatih/color"
    "github.com/google/uuid"
    "github.com/StupidWeasel/bootdev-blog-aggregator/gator/internal/database"
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

    result, err := s.db.GetUser(s.context, cmd.args[0])
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

    result, err := s.db.CreateUser(s.context, cmd.args[0])
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

    err := s.db.ResetUsers(s.context)
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

    results, err := s.db.GetUsers(s.context)
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

func buildBatchs(s *state, feed ScrapedFeed) []database.CreatePostParams{

    batchSize := s.config.PostBatchSize
    numBatches := (len(feed.Feed.Channel.Item) + batchSize - 1) / batchSize
    output := make([]database.CreatePostParams, numBatches, numBatches)

    for batch := 0; batch < numBatches; batch++ {
            start := batch*batchSize
            end := min(start+batchSize, len(feed.Feed.Channel.Item))
            thesePosts := feed.Feed.Channel.Item[start:end]
            output[batch] = database.CreatePostParams{
                Titles:         []string{},
                Urls:           []string{},
                Descriptions:   []string{},
                PublishedAts:   []time.Time{},
                FeedIds:        []int64{},
            }
            seen := make(map[string]struct{})
            for _,post := range thesePosts{
                if _, exists := seen[post.Link]; !exists {
                    seen[post.Link] = struct{}{}
                    output[batch].Titles = append(output[batch].Titles, post.Title)
                    output[batch].Urls = append(output[batch].Urls, post.Link)
                    output[batch].Descriptions = append(output[batch].Descriptions, post.Description)
                    output[batch].FeedIds = append(output[batch].FeedIds, int64(feed.FeedID))

                    parsedTime, err := time.Parse(time.RFC1123Z, post.PubDate)
                    if err != nil {
                        parsedTime, err = time.Parse(time.RFC3339, post.PubDate)
                        if err != nil {
                            output[batch].PublishedAts = append(output[batch].PublishedAts, time.Time{})
                        }else{
                            output[batch].PublishedAts = append(output[batch].PublishedAts, parsedTime)
                        }
                    }else{
                        output[batch].PublishedAts = append(output[batch].PublishedAts, parsedTime)
                    }
                }
            }
    }
    return output
}

func handlerAgg(s *state, cmd command) error{
    if len(cmd.args) != 1{
        return errors.New("usage: agg {frequency of updates}")
    }
    duration, err := time.ParseDuration(cmd.args[0])
    if err != nil{
        return errors.New("Unable to parse frequency. Corrent formatting is '30m' or '1h'")
    }
    fmt.Printf("Collecting feeds every %s\n", duration)

    ticker := time.NewTicker(duration)
    defer ticker.Stop()

    for ; ; {
        select {
        case <-s.context.Done():
            return nil
        default:
            output, err := scrapeFeeds(s)
            if err != nil {
                fmt.Fprintf(os.Stderr, "%v\n", err)
                continue
            }
            fmt.Printf("[Agg] Creating batches for FeedID: %d (items total: %d)\n", output.FeedID, len(output.Feed.Channel.Item))
            batches := buildBatchs(s,output)
            for i,batch := range batches{
                fmt.Printf("[Agg] Processing batch %d of %d.\n", i+1, len(batches))

                numRows, err := s.db.CreatePost(s.context, batch)
                if err != nil{
                    fmt.Fprintf(os.Stderr, "[Agg] Error inserting batch: %v\n", err)
                    continue
                }
                fmt.Printf("[Agg] Inserted batch, total rows: %d\n[Agg] (Note, this excludes duplicates)\n", numRows)

            }
            fmt.Println("[Agg] All done")

            select {
            case <-s.context.Done():
                return nil
            case <-ticker.C:
                continue
            }
        }
    }

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
    result, err := addFeed(s, cmd.args[0], cmd.args[1], userID)
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
        fmt.Println("No followed feeds to list!")
        return nil
    }
    fmt.Println("You are following:")
    for _,feed := range feeds{
        fmt.Printf("%s (%s)\n", feed.FeedName, feed.Url)
    }
    return nil
}

func handlerRemoveFeedFollow(s *state, cmd command, userID uuid.UUID) error{

    if len(cmd.args) != 1{
        return errors.New("usage: unfollow {url}")
    }

    err := removeFeedFollow(s, userID, strings.ToLower(cmd.args[0]))
    if err != nil{
        return err
    }
    fmt.Println("Feed unfollowed.")
    return nil

}

func handlerBrowseNext(s *state, cmd command, userID uuid.UUID) error{

    if len(cmd.args)>1{
        return fmt.Errorf("usage: %s {optional limit}", cmd.name)
    }
    limit := 2
    if len(cmd.args)==1{
        newLimit, err := strconv.Atoi(cmd.args[0])
        if err != nil || newLimit < 1 || newLimit>10 {
            return errors.New("That's not a valid limit, range is 1-10")
        }
        limit = newLimit
    }

    params := database.GetPostsForUser_ForwardParams{
        UserID:     s.config.CurrentUserID,
        Limit:      int32(limit),
    }

    if s.config.CurrentCursor != nil && !s.config.CurrentCursor.CursorTime.IsZero() && s.config.CurrentCursor.CursorId != 0 {
        params.HasCursor = true
        params.CursorTime = sql.NullTime{
            Time: s.config.CurrentCursor.CursorTime,
            Valid: true,
        }
        params.CursorID = s.config.CurrentCursor.CursorId
    } else {
        params.HasCursor = false
    }

    posts, err := s.db.GetPostsForUser_Forward(s.context, params)
    if err != nil{
        return fmt.Errorf("Failed to get posts: %w", err)
    }
    if len(posts)==0{
        return errors.New("No results")
    }
    lastPost := posts[len(posts)-1]

    err = s.config.SetCursor(lastPost.PublishedAt.Time, lastPost.FeedID)
    if err != nil{
        return fmt.Errorf("Failed to save position: %w", err)
    }

    for i,post := range posts{
        fmt.Printf("\n=== Post %d ===\n", i+1)
        fmt.Printf("Title: %s\n", post.Title)
        fmt.Printf("Published: %s\n", post.PublishedAt.Time.Format("2006-01-02 15:04"))
        fmt.Printf("URL: %s\n", post.Url)
        if post.Description.Valid {
            fmt.Printf("Description: %s\n\n", post.Description.String)
        }
    }
    return nil
}

func handlerBrowseBack(s *state, cmd command, userID uuid.UUID) error{

    if len(cmd.args)>1{
        return fmt.Errorf("usage: %s {optional limit}", cmd.name)
    }
    limit := 2
    if len(cmd.args)==1{
        newLimit, err := strconv.Atoi(cmd.args[0])
        if err != nil || newLimit < 1 || newLimit>10 {
            return errors.New("That's not a valid limit, range is 1-10")
        }
        limit = newLimit
    }

    params := database.GetPostsForUser_BackwardParams{
        UserID:     s.config.CurrentUserID,
        Limit:      int32(limit),
    }

    if s.config.CurrentCursor != nil && !s.config.CurrentCursor.CursorTime.IsZero() && s.config.CurrentCursor.CursorId != 0 {
        params.HasCursor = true
        params.CursorTime = sql.NullTime{
            Time: s.config.CurrentCursor.CursorTime,
            Valid: true,
        }
        params.CursorID = s.config.CurrentCursor.CursorId
    } else {
        params.HasCursor = false
    }

    posts, err := s.db.GetPostsForUser_Backward(s.context, params)
    if err != nil{
        return fmt.Errorf("Failed to get posts: %w", err)
    }
    if len(posts)==0{
        return errors.New("No results")
    }
    firstPost := posts[0]

    err = s.config.SetCursor(firstPost.PublishedAt.Time, firstPost.FeedID)
    if err != nil{
        return fmt.Errorf("Failed to save position: %w", err)
    }

    for i,post := range posts{
        fmt.Printf("\n=== Post %d ===\n", i+1)
        fmt.Printf("Title: %s\n", post.Title)
        fmt.Printf("Published: %s\n", post.PublishedAt.Time.Format("2006-01-02 15:04"))
        fmt.Printf("URL: %s\n", post.Url)
        if post.Description.Valid {
            fmt.Printf("Description: %s\n\n", post.Description.String)
        }
    }
    return nil
}
package main

import(
    "errors"
    "fmt"
    "io"
    "net/http"
    "context"
    "html"
    "encoding/xml"
    "github.com/google/uuid"
    "github.com/StupidWeasel/bootdev-blog-aggregator/gator/internal/database"
)

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error){

    client := &http.Client{
        Transport: &http.Transport{},
    }
 
    req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
    if err != nil {
        return nil, fmt.Errorf("Unable to create request: %w", err)
    }
    req.Header.Set("User-Agent", "gator/v0.01")
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to %s: %w", feedURL, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == 404{
        return nil, fmt.Errorf("%s not found", feedURL)
    }
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("Status code %d returned", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("Failed to read response body: %w", err)
    }

    var rss RSSFeed
    err = xml.Unmarshal(body, &rss)
    if err != nil {
        return nil, fmt.Errorf("Failed to unmarshal XML/RSS: %w", err)
    }

    rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
    rss.Channel.Description = html.UnescapeString(rss.Channel.Description)

    for i,item := range rss.Channel.Item{
        rss.Channel.Item[i].Title = html.UnescapeString(stripHTML(item.Title))
        rss.Channel.Item[i].Description = html.UnescapeString(stripHTML(item.Description))
    }

    return &rss, nil
}

func addFeed(s *state, name, url string, user_id uuid.UUID) (database.Feed, error){

    params := database.AddFeedParams{
            Name: name,
            Url: url,
            UserID: user_id,
    }

    result, err := s.db.AddFeed(s.context,params)
    if err != nil{
        return database.Feed{}, fmt.Errorf("Failed to insert into database: %w", err)
    }
    return result, nil
}

func getFeeds(s *state) ([]database.GetFeedsRow, error){

    results, err := s.db.GetFeeds(s.context)
    if err != nil{
        return nil, fmt.Errorf("Failed to get feeds: %w", err)
    }
    return results, nil
}

func addFeedFollow(s *state, user uuid.UUID, feedID int32) (database.CreateFeedFollowRow, error){

    params := database.CreateFeedFollowParams{
        UserID: user,
        FeedID: feedID,
    }
     
    result, err := s.db.CreateFeedFollow(s.context,params)
    if err != nil {
        return database.CreateFeedFollowRow{}, err
    }

    return result, nil
}

func getUserFeedFollow(s *state, user uuid.UUID) ([]database.GetFeedFollowsRow, error){

    results, err := s.db.GetFeedFollows(s.context, user)
    if err != nil{
        return nil, fmt.Errorf("Failed to get user feeds: %w", err)
    }
    return results, nil
}

func getFeed(s *state, url string) (database.GetFeedRow, error){

    result, err := s.db.GetFeed(s.context, url)
    if err != nil{
        return database.GetFeedRow{}, fmt.Errorf("Failed to get feed: %w", err)
    }
    return result, nil
}

func removeFeedFollow(s *state, user uuid.UUID, url string) error{

    params := database.UnFeedFollowParams{
        Url: url,
        UserID: user,
    }
    numRows, err := s.db.UnFeedFollow(s.context, params)
    if err != nil{
        return fmt.Errorf("Failed to delete feed follow: %w", err)
    }
    if numRows==0{
        return errors.New("No followed feed found, check the url")
    }
    return nil
}

func scrapeFeeds(s *state) (ScrapedFeed, error){

    nextFeed, err := s.db.GetNextFeedToFetch(s.context)
    if err != nil{
        return ScrapedFeed{}, fmt.Errorf("Failed to GetNextFeedToFetch: %w", err)
    }

    rowCount, err := s.db.MarkFeedFetched(s.context, nextFeed.ID)
    if err != nil || rowCount==0{
        return ScrapedFeed{}, fmt.Errorf("Failed to MarkFeedFetched (feed id:%d): %w", nextFeed.ID, err)
    }

    feed, err := fetchFeed(s.context, nextFeed.Url)
    if err != nil{
        return ScrapedFeed{}, fmt.Errorf("Failed to fetchFeed (%s): %w", nextFeed.Url, err)
    }
    return ScrapedFeed{FeedID: nextFeed.ID, Feed: feed}, nil
}
package config

import (
    "os"
    "fmt"
    "path/filepath"
    "encoding/json"
    "io"
    "time"
    "github.com/google/uuid"
)
type Cursor struct {
    CursorTime  time.Time `json: "cursor_time"`
    CursorId    int32 `json:"cursor_id"`
}

type Config struct {
    ConfigFile      string `json:"config_file"`
    ConfigPath      string `json:"config_path"`
    DbURL           string `json:"db_url"`
    PostBatchSize   int `json:"post_batch_size"`
    CurrentUserName string `json:"current_user_name"`
    CurrentUserID   uuid.UUID `json: "current_user_id"`
    CurrentCursor   *Cursor `json:"current_cursor"`
}


func unmarshalJSON[T any](r io.Reader, v *T) error{
    decoder := json.NewDecoder(r)
    //decoder.DisallowUnknownFields()
    if err := decoder.Decode(v); err != nil {
        return fmt.Errorf("Unable to unmarshal JSON: %w", err)
    }
    return nil
}

func (c *Config)SetUser(userName string, userID uuid.UUID) error{
    c.CurrentUserName = userName
    c.CurrentUserID = userID
    data, err := json.MarshalIndent(c," ", " ")
    if err != nil {
        return fmt.Errorf("Error marshalling JSON: %w", err)
    }
    
    err = os.WriteFile(c.ConfigPath, data, 0644)
    if err != nil {
        return fmt.Errorf("Unable to write file: %w", err)
    }
    return nil
}

func (c *Config)UnsetUser(){
    c.CurrentUserName = ""
    c.CurrentUserID = uuid.UUID{}
    c.CurrentCursor = nil
}

func (c *Config)ReadConfig() error{
    homeDir, err := os.UserHomeDir()
    if err != nil{
        return fmt.Errorf("No homedir: %w", err) 
    }

    fullPath := filepath.Join(homeDir,c.ConfigFile)
    data, err := os.Open(fullPath)
    if err != nil{
        return fmt.Errorf("%s file can't be read: %w", c.ConfigFile, err) 
    }
    
    err = unmarshalJSON(data, &c)
    if err != nil{
        return err 
    }

    c.ConfigPath = fullPath
    return nil
}

func (c *Config)SetCursor(cursorTime time.Time, cursorId int32) error{

    c.CurrentCursor = &Cursor{ 
            CursorTime: cursorTime,
            CursorId: cursorId,
    }

    data, err := json.MarshalIndent(c," ", " ")
    if err != nil {
        return fmt.Errorf("Error marshalling JSON: %w", err)
    }
    
    err = os.WriteFile(c.ConfigPath, data, 0644)
    if err != nil {
        return fmt.Errorf("Unable to write file: %w", err)
    }
    return nil
}
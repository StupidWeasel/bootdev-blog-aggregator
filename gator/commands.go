package main

import(
    "strings"
    "errors"
)

func (c *commands) register(name string, f func(*state, command) error) error{

    name = strings.ToLower(strings.TrimSpace(name))

    if name == ""{
        return errors.New("Command name required") 
    }
    if strings.Contains(name, " "){
        return errors.New("Command name may not have spaces") 
    }

    if f == nil{
        return errors.New("Invalid nil command function")
    }

    if _, exists := c.handlers[name]; exists {
        return errors.New("Command already exists")
    }

    c.handlers[name] = f
    return nil
}
func (c *commands) run(s *state, cmd command) error{

    if _, exists := c.handlers[cmd.name]; !exists {
        return errors.New("Command does not exist")
    }
    return c.handlers[cmd.name](s, cmd)
}

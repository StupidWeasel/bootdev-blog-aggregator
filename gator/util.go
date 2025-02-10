package main

import (
    "strings"
    "golang.org/x/net/html"
)

func stripHTML(s string) string {
    doc, err := html.Parse(strings.NewReader(s))
    if err != nil {
        return s
    }
    
    var textContent string
    var extractText func(*html.Node)
    extractText = func(n *html.Node) {
        if n.Type == html.TextNode {
            textContent += n.Data
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            extractText(c)
        }
    }
    extractText(doc)
    return strings.Join(strings.Fields(textContent), " ")
}


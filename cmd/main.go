package main

import (
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	url := "https://webscraper.io/test-sites/e-commerce/allinone"

	log.Println("Visiting", url)

	doc, err := fetch(url)
	if err != nil {
		log.Fatal(err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			trimmed := strings.TrimSpace(n.FirstChild.Data)
			if trimmed == "" {
				return
			}

			log.Println("Found link:", trimmed)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

func fetch(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

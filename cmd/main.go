package main

import (
	"log"
	"net/http"

	"golang.org/x/net/html"
)

func main() {
	url := "https://monzo.com"

	log.Println("Visiting", url)

	doc, err := fetch(url)
	if err != nil {
		log.Fatal(err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		//TODO: Use a data atom here to find links instead
		if n.Type == html.ElementNode && n.Data == "a" {
			attrs := n.Attr
			for _, a := range attrs {
				if a.Key != "href" {
					continue
				}

				log.Println("Found link:", a.Val)
			}

		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

func visitLinks(

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

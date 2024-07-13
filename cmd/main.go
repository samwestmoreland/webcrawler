package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	startURL := "https://monzo.com"
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Visiting", startURL)
	log.Println("Host:", parsedURL.Host)

	doc, err := fetch(startURL)
	if err != nil {
		log.Fatal(err)
	}

	links, err := extractLinks(doc, parsedURL.Host)
	if err != nil {
		log.Fatal(err)
	}

	for _, link := range links {
		log.Println(link)
	}
}

func extractLinks(doc *html.Node, host string) ([]string, error) {
	var links []string
	var (
		invalidLinksCount int
		validLinksCount   int
		erroredCount      int
	)
	seen := make(map[string]struct{})

	var f func(*html.Node)
	f = func(n *html.Node) {
		//TODO: Use a data atom here to find links instead
		if n.Type == html.ElementNode && n.Data == "a" {
			attrs := n.Attr
			for _, a := range attrs {
				if a.Key != "href" {
					continue
				}

				u, err := url.Parse(a.Val)
				if err == nil && u.Host == host {
					if _, ok := seen[a.Val]; ok {
						continue
					}
					seen[a.Val] = struct{}{}
					links = append(links, a.Val)
					validLinksCount++

				} else if err == nil && u.Host == "" && strings.HasPrefix(a.Val, "/") {
					// This is a relative link
					if _, ok := seen[a.Val]; ok {
						continue
					}
					seen[a.Val] = struct{}{}
					links = append(links, a.Val)
					validLinksCount++
				} else if err == nil {
					log.Printf("Invalid URL: %q u.Host: %q", a.Val, u.Host)
					invalidLinksCount++
				} else {
					log.Println("Invalid URL:", a.Val)
					erroredCount++
				}
			}

		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	log.Println("Invalid links count:", invalidLinksCount)
	log.Println("Valid links count:", validLinksCount)
	log.Println("Errored count:", erroredCount)

	return links, nil
}

func fetch(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

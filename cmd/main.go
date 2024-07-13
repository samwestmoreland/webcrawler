package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/net/html"

	surl "github.com/samwestmoreland/webcrawler/src/url"
)

func main() {
	startURL := "https://www.monzo.com"

	crawler, err := NewCrawler(startURL)
	if err != nil {
		log.Fatal(err)
	}
	crawler.Crawl()

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

	log.Println("Calling extractLinks with host", parsedURL.Host)
	links, err := extractLinks(doc, parsedURL.Scheme+"://"+parsedURL.Host)
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
				if err == nil && (u.Host == host || u.Host == "") {
					if _, ok := seen[a.Val]; ok {
						continue
					}
					seen[a.Val] = struct{}{}
					normalisedURL, err := surl.Normalise(host, a.Val)
					if err != nil {
						log.Printf("Failed to normalise: %q, err: %v", a.Val, err)
						erroredCount++
						continue
					} else {
						log.Println("Normalised:", normalisedURL)
					}
					links = append(links, normalisedURL)
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

// isValidURL checks if the url is part of the same subdomain if it is absolute,
// otherwise it returns true if the url is relative. If the url is not parsable,
// it returns false
func isValidURL(url string) bool {
	u, err := url.Parse(url)
	if err != nil {
		return false
	}

	normalisedURL, err := surl.Normalise(u.Host, url)
	if err != nil {
		return false
	}
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

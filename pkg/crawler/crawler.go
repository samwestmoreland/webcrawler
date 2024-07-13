package crawler

import (
	"fmt"
	"log"
	"net/http"

	surl "github.com/samwestmoreland/webcrawler/pkg/url"
	"golang.org/x/net/html"
)

type Crawler struct {
	// this would be monzo.com or community.monzo.com for example
	subdomain string

	links []string
}

func NewCrawler(url string) (*Crawler, error) {
	u, err := surl.Parse(url)
	if err != nil {
		return nil, err
	}
	subdomain := u.Host

	return &Crawler{subdomain: subdomain}, nil
}

func (c Crawler) Crawl() {
}

func (c Crawler) extractLinks(doc *html.Node) ([]string, error) {
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

				if a.Val == "" {
					continue
				}

				if _, ok := seen[a.Val]; ok {
					continue
				}

				seen[a.Val] = struct{}{}

				if !c.isValidURL(a.Val) {
					invalidLinksCount++
					continue
				}

				links = append(links, a.Val)
				validLinksCount++
				if validLinksCount%100 == 0 {
					log.Println("Valid links count:", validLinksCount)
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

func (c Crawler) fetch(url string) (*html.Node, error) {
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

// isValidURL checks if the url is part of the same subdomain if it is absolute,
// otherwise it returns true if the url is relative. If the url is not parsable,
// it returns false
func (c Crawler) isValidURL(url string) bool {
	u, err := surl.Parse(url)
	if err != nil {
		return false
	}

	normalisedURL, err := surl.Normalise(c.subdomain, url)
	if err != nil {
		return false
	}

	return u.Hostname() == normalisedURL.Host
}

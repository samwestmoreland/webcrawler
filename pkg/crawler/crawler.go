package crawler

import (
	"log"
	"net/url"

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

	return Crawler{subdomain: subdomain}
}

func (c Crawler) Crawl() {
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

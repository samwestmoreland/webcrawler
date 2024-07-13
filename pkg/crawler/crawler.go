package crawler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/samwestmoreland/webcrawler/pkg/url"
	"golang.org/x/net/html"
)

type Crawler struct {
	// this would be www.monzo.com or www.community.monzo.com for example
	subdomain string

	// this would be https://monzo.com or https://www.community.monzo.com
	url string

	// the links we found
	links []string
}

func NewCrawler(u string) (*Crawler, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	return &Crawler{
		subdomain: parsed.Subdomain,
		url:       u,
	}, nil
}

func (c Crawler) Crawl() error {
	log.Println("Crawling", c.subdomain)

	doc, err := c.fetch(c.url)
	if err != nil {
		return fmt.Errorf("error fetching %s: %w", c.subdomain, err)
	}

	links, err := c.extractLinks(doc)
	if err != nil {
		return fmt.Errorf("error extracting links: %w", err)
	}

	c.links = links

	return nil
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
					log.Println("Invalid link:", a.Val)
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

// isValidURL checks if an href is a valid url, is part of the same subdomain if
// it is absolute. If the URL is a relative path it will return true. If the url
// is not parsable, it returns false
func (c Crawler) isValidURL(href string) bool {
	if _, err := url.Parse(href); err != nil {
		log.Println("Couldn't parse href:", href)
		return false
	}

	normalisedURL, err := url.Normalise(c.subdomain, href)
	if err != nil {
		log.Println("Couldn't normalise href:", href)
		return false
	}

	isSameSubdomain := url.IsSameSubdomain(c.subdomain, normalisedURL.URL)
	if !isSameSubdomain {
		log.Printf("Href is not in the same subdomain. subdomain: %s, href: %s\n", c.subdomain, normalisedURL.URL)
		return false
	}

	return true
}

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
	host string

	// this would be https://monzo.com or https://www.community.monzo.com
	url string

	// the links we found
	links map[string]struct{}
}

func NewCrawler(u string) (*Crawler, error) {
	parsed, err := url.ParseURLString(u)
	if err != nil {
		return nil, err
	}

	return &Crawler{
		host: parsed.Host,
		url:  u,
	}, nil
}

// Crawl performs a BFS traversal of the domain
func (c Crawler) Crawl() error {
	log.Println("Crawling", c.host)

	queue := []string{c.url}
	visitedSet := make(map[string]struct{})

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if _, visited := visitedSet[current]; visited {
			continue
		}
		visitedSet[current] = struct{}{}

		doc, err := c.fetch(current)
		if err != nil {
			return fmt.Errorf("error fetching %s: %w", current, err)
		}

		links, err := c.extractLinks(doc)
		if err != nil {
			return fmt.Errorf("error extracting links: %w", err)
		}

		for _, link := range links {
			// validation is done in extractLinks() so we can safely add to the
			// queue here
			queue = append(queue, link)
		}

	}

	c.links = visitedSet

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

				normalised, err := url.Normalise(c.host, a.Val)
				if err != nil {
					erroredCount++
					continue
				}

				if _, ok := seen[normalised.URL]; ok {
					continue
				}
				seen[a.Val] = struct{}{}

				if !c.isValidURL(normalised) {
					log.Println("Not in subdomain:", normalised.URL)
					invalidLinksCount++
					continue
				}

				links = append(links, normalised.URL)
				validLinksCount++
				log.Println("Found link:", normalised.URL)
				// if validLinksCount%100 == 0 {
				// 	log.Println("Valid links count:", validLinksCount)
				// }
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

// fetch performs an HTTP GET request. It expects a fully qualified URL
// to be passed in, i.e. one with a scheme and hostname
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

func (c Crawler) isValidURL(u *url.URL) bool {
	same, err := url.IsSameHost(c.host, u.Host)

	return err == nil && same
}

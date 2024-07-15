package crawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/samwestmoreland/webcrawler/pkg/url"
	"golang.org/x/net/html"
)

type Results struct {
	Links         []string
	ExternalLinks []string
	ErroredLinks  []string
}

type Crawler struct {
	// www.foo.com or www.subdomain.foo.com, for example. Used for comparisons.
	host string

	// https://foo.com or https://www.subdomain.foo.com, for example. A
	// visitable URL.
	url string

	// the links found
	results Results

	seen map[string]struct{}

	// log file
	logFile io.Writer
}

// NewCrawler creates a new Crawler
func NewCrawler(u string) (*Crawler, error) {
	parsed, err := url.ParseURLString(u)
	if err != nil {
		return nil, err
	}

	return &Crawler{
		host: parsed.Host,
		url:  u,
		seen: make(map[string]struct{}),
		// write to stdout by default
		logFile: os.Stdout,
	}, nil
}

func NewCrawlerWithLogFile(u string, logFile io.Writer) (*Crawler, error) {
	crawler, err := NewCrawler(u)
	if err != nil {
		return nil, err
	}

	crawler.logFile = logFile
	return crawler, nil
}

func (c *Crawler) log(msg string) {
	msg = fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
	if _, err := c.logFile.Write([]byte(msg)); err != nil {
		log.Println(err)
	}
}

func (c *Crawler) OutputResults() {
	// Write the links to the log file
	c.logFile.Write([]byte(fmt.Sprintf("links found: %d\n", len(c.results.Links))))
	for _, link := range c.results.Links {
		if _, err := c.logFile.Write([]byte(link + "\n")); err != nil {
			log.Println(err)
		}
	}
	c.logFile.Write([]byte("\n"))

	c.logFile.Write([]byte(fmt.Sprintf("external links found: %d\n", len(c.results.ExternalLinks))))
	for _, link := range c.results.ExternalLinks {
		if _, err := c.logFile.Write([]byte(link + "\n")); err != nil {
			log.Println(err)
		}
	}
	c.logFile.Write([]byte("\n"))

	c.logFile.Write([]byte(fmt.Sprintf("errored links found: %d\n", len(c.results.ErroredLinks))))
	for _, link := range c.results.ErroredLinks {
		if _, err := c.logFile.Write([]byte(link + "\n")); err != nil {
			log.Println(err)
		}
	}

	// Print totals to stdout
	fmt.Printf("links found: %d\n", len(c.results.Links))
	fmt.Printf("external links found: %d\n", len(c.results.ExternalLinks))
	fmt.Printf("errored links found: %d\n", len(c.results.ErroredLinks))
}

// Crawl performs a BFS traversal of the domain
func (c *Crawler) Crawl() error {
	c.log(fmt.Sprintf("Crawling %s\n", c.url))
	if c.logFile != os.Stdout {
		fmt.Printf("Crawling %s\n", c.url)
	}

	queue := []string{c.url}
	visitedSet := make(map[string]struct{})

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		c.log(fmt.Sprintf("Visiting %s\n", current))

		if _, visited := visitedSet[current]; visited {
			c.log(fmt.Sprintf("Already visited %s\n", current))
			continue
		}
		visitedSet[current] = struct{}{}
		c.results.Links = append(c.results.Links, current)

		doc, err := c.fetch(current)
		if err != nil {
			c.results.ErroredLinks = append(c.results.ErroredLinks, current)
			continue
		}

		c.results.Links = append(c.results.Links, current)

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

	return nil
}

func (c *Crawler) extractLinks(doc *html.Node) ([]string, error) {
	var links []string

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
					c.results.ErroredLinks = append(c.results.ErroredLinks, a.Val)
					continue
				}

				if _, ok := c.seen[normalised.URL]; ok {
					continue
				}
				c.seen[normalised.URL] = struct{}{}

				if !c.isValidURL(normalised) {
					c.results.ExternalLinks = append(c.results.ExternalLinks, normalised.URL)
					continue
				}

				links = append(links, normalised.URL)
			}

		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return links, nil
}

// fetch performs an HTTP GET request. It expects a fully qualified URL
// to be passed in, i.e. one with a scheme and hostname
func (c *Crawler) fetch(urlToFetch string) (*html.Node, error) {
	u, err := url.ParseURLString(urlToFetch)
	if err != nil {
		return nil, fmt.Errorf("error parsing %q for fetch: %s", urlToFetch, err)
	}

	resp, err := http.Get(u.URL)
	if err != nil {
		return nil, fmt.Errorf("error getting %q: %s", urlToFetch, err)
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

func (c *Crawler) isValidURL(u *url.URL) bool {
	//TODO: return an error here as well
	same, err := url.IsSameHost(c.host, u.Host)

	return err == nil && same
}

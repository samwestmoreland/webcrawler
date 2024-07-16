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

const (
	defaultMaxRetriesOnStatusAccepted    = 5
	defaultStatusAcceptedPollingInterval = 5 * time.Second
	defaultRequestTimeout                = 5 * time.Second
	defaultLogFileName                   = "crawler.log"
)

type erroredLink struct {
	url      string
	errorMsg string
}

type Results struct {
	Links         []string
	ExternalLinks []string
	ErroredLinks  []erroredLink
	TotalTime     time.Duration
}

type Crawler struct {
	// https://foo.com or https://www.subdomain.foo.com, for example. A
	// visitable URL.
	StartURL string
	// timeout to set on the http client
	RequestTimeout time.Duration
	// the links found
	Results     Results
	LogFileName string

	// www.foo.com or www.subdomain.foo.com, for example. Used for comparisons.
	host string
	// links we've already seen
	seen map[string]struct{}
	// log file
	logFile io.Writer

	client *http.Client

	// retry parameters on status code 202
	statusAcceptedMaxRetries      int
	statusAcceptedPollingInterval time.Duration
}

// NewCrawler creates a new Crawler
func NewDefaultCrawler(u string) (*Crawler, error) {
	parsed, err := url.ParseURLString(u)
	if err != nil {
		return nil, err
	}

	var httpClient = &http.Client{
		Timeout: defaultRequestTimeout,
	}

	return &Crawler{
		StartURL:                      u,
		LogFileName:                   defaultLogFileName,
		RequestTimeout:                defaultRequestTimeout,
		host:                          parsed.Host,
		seen:                          make(map[string]struct{}),
		client:                        httpClient,
		statusAcceptedMaxRetries:      defaultMaxRetriesOnStatusAccepted,
		statusAcceptedPollingInterval: defaultStatusAcceptedPollingInterval,
	}, nil
}

func (c *Crawler) log(msg string) {
	msg = fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
	if _, err := c.logFile.Write([]byte(msg)); err != nil {
		log.Println(err)
	}
}

func (c *Crawler) write(msg string) {
	if _, err := c.logFile.Write([]byte(msg)); err != nil {
		log.Println(err)
	}
}

func (c *Crawler) OutputResults() {
	// Write the links to the log file
	c.write(fmt.Sprintf("links found: %d\n", len(c.Results.Links)))
	for _, link := range c.Results.Links {
		c.write(link + "\n")
	}

	c.write("\n")

	c.write(fmt.Sprintf("external links found: %d\n", len(c.Results.ExternalLinks)))
	for _, link := range c.Results.ExternalLinks {
		c.write(link + "\n")
	}

	c.write("\n")

	c.write(fmt.Sprintf("errored links found: %d\n", len(c.Results.ErroredLinks)))
	for _, link := range c.Results.ErroredLinks {
		c.write(fmt.Sprintf("%s: %s\n", link.url, link.errorMsg))
	}

	c.write(fmt.Sprintf("crawling took %.2f seconds\n", c.Results.TotalTime.Seconds()))

	// Print totals to stdout
	fmt.Printf("links found: %d\n", len(c.Results.Links))
	fmt.Printf("external links found: %d\n", len(c.Results.ExternalLinks))
	fmt.Printf("links that errorred: %d\n", len(c.Results.ErroredLinks))
	fmt.Printf("crawling took %.2f seconds\n", c.Results.TotalTime.Seconds())
	fmt.Printf("please see %s for Results\n", c.LogFileName)
}

// Crawl does some setup and then starts the crawl
func (c *Crawler) Crawl() error {
	// initialise the log file
	logFile, err := os.OpenFile(c.LogFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %s", err)
	}

	c.logFile = logFile

	c.log(fmt.Sprintf("crawling %s\n", c.StartURL))
	fmt.Printf("Crawling %s\n", c.StartURL)

	start := time.Now()
	defer func() {
		c.Results.TotalTime = time.Since(start)
	}()

	return c.crawl(c.StartURL)
}

// crawl performs a BFS traversal of the domain
func (c *Crawler) crawl(u string) error {
	queue := []string{u}
	visitedSet := make(map[string]struct{})

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		fetchableURL, err := url.ParseURLString(current)
		if err != nil {
			c.Results.ErroredLinks = append(c.Results.ErroredLinks, erroredLink{url: current, errorMsg: err.Error()})
			continue
		}

		if _, visited := visitedSet[fetchableURL.URL]; visited {
			c.log(fmt.Sprintf("Already visited %s\n", fetchableURL.URL))
			continue
		}
		visitedSet[fetchableURL.URL] = struct{}{}

		c.log(fmt.Sprintf("visiting %s\n", fetchableURL.URL))

		doc, err := c.fetch(fetchableURL.URL)
		if err != nil {
			c.Results.ErroredLinks = append(c.Results.ErroredLinks, erroredLink{url: fetchableURL.URL, errorMsg: err.Error()})
			continue
		}

		// Add the parsed URL to results slice. It's been normalised so this
		// should avoid duplicates
		c.Results.Links = append(c.Results.Links, fetchableURL.URL)

		links, err := c.extractLinks(doc)
		if err != nil {
			return fmt.Errorf("error extracting links: %w", err)
		}

		for _, link := range links {
			// validation is done in extractLinks() _and_ before we fetch, so
			// we can safely just add to the queue here
			queue = append(queue, link)
		}

		// Log every 100 visited pages so we know we're making progress
		if len(visitedSet)%100 == 0 {
			c.log(fmt.Sprintf("visited %d pages\n", len(visitedSet)))

			if c.logFile != os.Stdout {
				fmt.Printf("visited %d pages\n", len(visitedSet))
			}
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
					c.Results.ErroredLinks = append(c.Results.ErroredLinks, erroredLink{url: a.Val, errorMsg: err.Error()})
					continue
				}

				if _, ok := c.seen[normalised.URL]; ok {
					continue
				}
				c.seen[normalised.URL] = struct{}{}

				if !c.isValidURL(normalised) {
					c.Results.ExternalLinks = append(c.Results.ExternalLinks, normalised.URL)
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
	poll := func(urlToFetch string) (*html.Node, error) {
		for retries := 0; retries < c.statusAcceptedMaxRetries; retries++ {
			resp, err := c.client.Get(urlToFetch)
			if err != nil {
				return nil, fmt.Errorf("error getting %q: %s", urlToFetch, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusAccepted {
				c.log(fmt.Sprintf("status code %d, sleeping for %.2f seconds before retrying\n",
					resp.StatusCode,
					c.statusAcceptedPollingInterval.Seconds()))

				time.Sleep(c.statusAcceptedPollingInterval)

				continue
			}

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
			}

			doc, err := html.Parse(resp.Body)
			if err != nil {
				return nil, err
			}

			return doc, nil
		}

		return nil, fmt.Errorf("failed to fetch %q after %d retries", urlToFetch, c.statusAcceptedMaxRetries)
	}

	resp, err := c.client.Get(urlToFetch)
	if err != nil {
		return nil, fmt.Errorf("error getting %q: %s", urlToFetch, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return poll(urlToFetch)
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

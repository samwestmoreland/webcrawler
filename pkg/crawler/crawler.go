package crawler

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/samwestmoreland/webcrawler/pkg/url"
	"golang.org/x/net/html"
)

const (
	defaultMaxRetriesOnStatusAccepted    = 5
	defaultStatusAcceptedPollingInterval = 5 * time.Second
	defaultRequestTimeout                = 5 * time.Second
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
	StartURL    *url.URL // the URL with which to start the crawl
	Results     Results  // a Results struct containing the results of the crawl
	LogFileName string   // the name of the log file

	// www.foo.com or www.subdomain.foo.com, for example. Used for comparisons.
	Host string

	seen map[string]struct{} // for keeping track of visited URLs

	// client
	client         *http.Client
	requestTimeout time.Duration // timeout to set on the http client

	// output files
	logger      *log.Logger
	resultsFile io.Writer

	// retry parameters on status code 202
	statusAcceptedMaxRetries      int
	statusAcceptedPollingInterval time.Duration
}

// NewCrawler creates a new Crawler
func NewCrawler(u string, logger *log.Logger, resultsFile io.Writer) (*Crawler, error) {
	parsed, err := url.ParseURLString(u, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %s", err)
	}

	// Ensure that the URL has a scheme because we'll use it later for normalising relative path URLs
	if parsed.Scheme == "" {
		return nil, fmt.Errorf("url must have a scheme (e.g. https): %s", u)
	}

	var httpClient = &http.Client{
		Timeout: defaultRequestTimeout,
	}

	return &Crawler{
		StartURL:                      parsed,
		logger:                        logger,
		resultsFile:                   resultsFile,
		requestTimeout:                defaultRequestTimeout,
		Host:                          parsed.Host,
		seen:                          make(map[string]struct{}),
		client:                        httpClient,
		statusAcceptedMaxRetries:      defaultMaxRetriesOnStatusAccepted,
		statusAcceptedPollingInterval: defaultStatusAcceptedPollingInterval,
	}, nil
}

// NewCrawlerDiscardOutput creates a new Crawler with no output. Used for testing
func NewCrawlerDiscardOutput(u string) (*Crawler, error) {
	return NewCrawler(u, log.New(ioutil.Discard, "", 0), ioutil.Discard)
}

// OutputResults writes the results to the output file
func (c *Crawler) OutputResults() {
	// Write the links to the log file
	c.resultsFile.Write([]byte(fmt.Sprintf("links found: %d\n", len(c.Results.Links))))
	for _, link := range c.Results.Links {
		c.resultsFile.Write([]byte(fmt.Sprintf("%s\n", link)))
	}

	c.resultsFile.Write([]byte("\n"))

	c.resultsFile.Write([]byte(fmt.Sprintf("external links found: %d\n", len(c.Results.ExternalLinks))))
	for _, link := range c.Results.ExternalLinks {
		c.resultsFile.Write([]byte(fmt.Sprintf("%s\n", link)))
	}

	c.resultsFile.Write([]byte("\n"))

	c.resultsFile.Write([]byte(fmt.Sprintf("errored links found: %d\n", len(c.Results.ErroredLinks))))
	for _, link := range c.Results.ErroredLinks {
		c.resultsFile.Write([]byte(fmt.Sprintf("%s: %s\n", link.url, link.errorMsg)))
	}

	// Print stats to stdout
	fmt.Println()
	fmt.Println("Crawler stats:")
	fmt.Println("-------------")
	fmt.Printf("links found:          %d\n", len(c.Results.Links))
	fmt.Printf("external links found: %d\n", len(c.Results.ExternalLinks))
	fmt.Printf("links that errorred:  %d\n", len(c.Results.ErroredLinks))
	fmt.Printf("crawling took %.2f seconds\n", c.Results.TotalTime.Seconds())
}

// Crawl does some setup and then starts the crawl
func (c *Crawler) Crawl() error {
	c.logger.Printf("crawling %s\n", c.StartURL)
	fmt.Printf("Crawling %s\n", c.StartURL)

	start := time.Now()
	defer func() {
		c.Results.TotalTime = time.Since(start)
	}()

	return c.crawl(c.StartURL)
}

// crawl performs a BFS traversal of the domain
func (c *Crawler) crawl(u *url.URL) error {
	fmt.Println("In crawl")
	queue := []string{u.URL}
	visitedSet := make(map[string]struct{})

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		fetchableURL, err := url.ParseURLString(current, c.StartURL.Scheme)
		if err != nil {
			c.Results.ErroredLinks = append(c.Results.ErroredLinks, erroredLink{url: current, errorMsg: err.Error()})
			continue
		}

		if _, visited := visitedSet[fetchableURL.URL]; visited {
			c.logger.Printf("already visited %s\n", fetchableURL.URL)
			continue
		}
		visitedSet[fetchableURL.URL] = struct{}{}

		c.logger.Printf("visiting %s\n", fetchableURL.URL)
		fmt.Println("visiting", fetchableURL.URL)

		doc, err := c.Fetch(fetchableURL.URL)
		if err != nil {
			c.Results.ErroredLinks = append(c.Results.ErroredLinks, erroredLink{url: fetchableURL.URL, errorMsg: err.Error()})
			fmt.Printf("error while fetching %s: %s\n", fetchableURL.URL, err)
			continue
		}

		// Add the parsed URL to results slice.
		c.Results.Links = append(c.Results.Links, fetchableURL.URL)

		links, err := c.ExtractLinks(doc)
		if err != nil {
			return fmt.Errorf("error extracting links: %w", err)
		}

		for _, link := range links {
			// validation is done in ExtractLinks() _and_ before we fetch, so
			// we can safely just add to the queue here
			queue = append(queue, link)
		}

		// Log every 100 visited pages so the user knows progress is being made
		if len(visitedSet)%100 == 0 {
			c.logger.Printf("visited %d pages\n", len(visitedSet))
			fmt.Printf("visited %d pages\n", len(visitedSet))
		}
	}

	return nil
}

// ExtractLinks extracts links from the HTML document
func (c *Crawler) ExtractLinks(doc *html.Node) ([]string, error) {
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

				normalised, err := url.Normalise(c.Host, a.Val)
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

// Fetch performs an HTTP GET request. It expects a fully qualified URL
// to be passed in, i.e. one with a scheme and hostname
func (c *Crawler) Fetch(urlToFetch string) (*html.Node, error) {
	poll := func(urlToFetch string) (*html.Node, error) {
		for retries := 0; retries < c.statusAcceptedMaxRetries; retries++ {
			resp, err := c.client.Get(urlToFetch)
			if err != nil {
				return nil, fmt.Errorf("error getting %q: %s", urlToFetch, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusAccepted {
				c.logger.Printf("status code %d, sleeping for %.0f seconds before retrying\n",
					resp.StatusCode,
					c.statusAcceptedPollingInterval.Seconds())

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
	same, err := url.IsSameHost(c.Host, u.Host)

	return err == nil && same
}

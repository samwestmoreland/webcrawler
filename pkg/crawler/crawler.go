package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	tickerInterval                       = 2 * time.Second
)

// Errors.
var (
	ErrMaxRetriesReached = errors.New("max retries reached")
	ErrBadStatusCode     = errors.New("got bad response code")
	ErrURLMissingScheme  = errors.New("url missing scheme")
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
	StartURL *url.URL // the URL with which to start the crawl
	Results  Results  // a Results struct containing the results of the crawl

	// www.foo.com or www.subdomain.foo.com, for example. Used for comparisons.
	Host string

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

// NewCrawler creates a new Crawler.
func NewCrawler(u string, logger *log.Logger, resultsFile io.Writer) (*Crawler, error) {
	parsed, err := url.ParseURLString(u, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	// Ensure that the URL has a scheme because we'll use it later for normalising relative path URLs
	if parsed.Scheme == "" {
		return nil, fmt.Errorf("url must have a scheme (e.g. https): %s: %w", u, ErrURLMissingScheme)
	}

	httpClient := &http.Client{
		Timeout: defaultRequestTimeout,
	}

	results := Results{
		Links:         []string{},
		ExternalLinks: []string{},
		ErroredLinks:  []erroredLink{},
		TotalTime:     0,
	}

	return &Crawler{
		StartURL:                      parsed,
		logger:                        logger,
		Results:                       results,
		resultsFile:                   resultsFile,
		requestTimeout:                defaultRequestTimeout,
		Host:                          parsed.Host,
		client:                        httpClient,
		statusAcceptedMaxRetries:      defaultMaxRetriesOnStatusAccepted,
		statusAcceptedPollingInterval: defaultStatusAcceptedPollingInterval,
	}, nil
}

// NewCrawlerDiscardOutput creates a new Crawler with no output. Used for testing.
func NewCrawlerDiscardOutput(u string) (*Crawler, error) {
	return NewCrawler(u, log.New(io.Discard, "", 0), io.Discard)
}

// Crawl does some setup and then starts the crawl.
func (c *Crawler) Crawl() error {
	c.logger.Printf("crawling %s\n", c.StartURL.URL)
	fmt.Printf("Crawling %s\n", c.StartURL.URL) //nolint:forbidigo

	start := time.Now()
	defer func() {
		c.Results.TotalTime = time.Since(start)
	}()

	return c.crawl(c.StartURL)
}

// crawl performs a BFS traversal of the domain.
func (c *Crawler) crawl(u *url.URL) error {
	queue := []string{u.URL}
	visitedSet := make(map[string]struct{})

	// Periodically output number of pages visited
	ticker := time.NewTicker(tickerInterval)
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				fmt.Printf("visited %d pages\n", len(visitedSet)) //nolint:forbidigo
			}
		}
	}()

	defer close(stop)

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

		doc, err := c.Fetch(fetchableURL.URL)
		if err != nil {
			c.Results.ErroredLinks = append(c.Results.ErroredLinks, erroredLink{url: fetchableURL.URL, errorMsg: err.Error()})
			c.logger.Printf("error fetching %s: %s\n", fetchableURL.URL, err)

			continue
		}

		// Add the parsed URL to results slice.
		c.Results.Links = append(c.Results.Links, fetchableURL.URL)

		links, err := c.ExtractLinks(doc)
		if err != nil {
			return fmt.Errorf("error extracting links: %w", err)
		}

		// validation is done in ExtractLinks() _and_ before we fetch, so
		// we can safely just add to the queue here
		queue = append(queue, links...)
	}

	return nil
}

// ExtractLinks extracts links from the HTML document.
func (c *Crawler) ExtractLinks(doc *html.Node) ([]string, error) {
	var links []string

	seen := make(map[string]struct{})

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			attrs := n.Attr
			for _, a := range attrs {
				if a.Key != "href" || a.Val == "" {
					continue
				}

				resolved, err := url.ResolvePath(c.Host, a.Val)
				if err != nil {
					c.Results.ErroredLinks = append(c.Results.ErroredLinks, erroredLink{url: a.Val, errorMsg: err.Error()})

					continue
				}

				if _, ok := seen[resolved.URL]; ok {
					continue
				}

				seen[resolved.URL] = struct{}{}

				if !c.isInternal(resolved) {
					c.Results.ExternalLinks = append(c.Results.ExternalLinks, resolved.URL)

					continue
				}

				links = append(links, resolved.URL)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return links, nil
}

func (c *Crawler) doGetWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	return resp, nil
}

// Fetch performs an HTTP GET request. It expects a fully qualified URL
// to be passed in, i.e. one with a scheme and hostname.
func (c *Crawler) Fetch(urlToFetch string) (*html.Node, error) {
	poll := func(urlToFetch string) (*html.Node, error) {
		for range c.statusAcceptedMaxRetries {
			resp, err := c.doGetWithContext(context.Background(), urlToFetch)
			if err != nil {
				return nil, fmt.Errorf("error getting %q: %w", urlToFetch, err)
			}
			defer resp.Body.Close()

			// Check the status code
			if resp.StatusCode == http.StatusAccepted {
				c.logger.Printf("status code %d, sleeping for %.0f seconds before retrying\n",
					resp.StatusCode,
					c.statusAcceptedPollingInterval.Seconds())

				time.Sleep(c.statusAcceptedPollingInterval)

				continue
			}

			if resp.StatusCode != http.StatusOK {
				return nil, ErrBadStatusCode
			}

			doc, err := html.Parse(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error parsing %q: %w", urlToFetch, err)
			}

			return doc, nil
		}

		return nil, fmt.Errorf(
			"failed to fetch %q after %d retries: %w",
			urlToFetch,
			c.statusAcceptedMaxRetries,
			ErrMaxRetriesReached,
		)
	}

	resp, err := c.doGetWithContext(context.Background(), urlToFetch)
	if err != nil {
		return nil, fmt.Errorf("error getting %q: %w", urlToFetch, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return poll(urlToFetch)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing html from %q: %w", urlToFetch, err)
	}

	return doc, nil
}

// isInternal returns false if the URL is not internal to the crawler's host
// _or_ if there is an error trying to figure it out.
func (c *Crawler) isInternal(u *url.URL) bool {
	same, err := url.IsSameHost(c.Host, u.Host)

	return err == nil && same
}

// OutputResults writes the results to the output file.
func (c *Crawler) OutputResults() {
	if _, err := c.resultsFile.Write(
		[]byte(fmt.Sprintf(
			"links found: %d\n", len(c.Results.Links)))); err != nil {
		c.logger.Printf("error writing to results file: %v", err)
	}

	for _, link := range c.Results.Links {
		if _, err := c.resultsFile.Write([]byte(link + "\n")); err != nil {
			c.logger.Printf("error writing to results file: %v", err)
		}
	}

	if _, err := c.resultsFile.Write(
		[]byte(fmt.Sprintf("\nexternal links found: %d\n",
			len(c.Results.ExternalLinks)))); err != nil {
		c.logger.Printf("error writing to results file: %v", err)
	}

	for _, link := range c.Results.ExternalLinks {
		if _, err := c.resultsFile.Write([]byte(link + "\n")); err != nil {
			c.logger.Printf("error writing to results file: %v", err)
		}
	}

	if _, err := c.resultsFile.Write(
		[]byte(fmt.Sprintf("\nerrored links found: %d\n",
			len(c.Results.ErroredLinks)))); err != nil {
		c.logger.Printf("error writing to results file: %v", err)
	}

	for _, link := range c.Results.ErroredLinks {
		if _, err := c.resultsFile.Write(
			[]byte(fmt.Sprintf("%s: %s\n", link.url, link.errorMsg))); err != nil {
			c.logger.Printf("error writing to results file: %v", err)
		}
	}

	// Print stats to stdout
	fmt.Println()                                                             //nolint:forbidigo
	fmt.Println("Crawler stats:")                                             //nolint:forbidigo
	fmt.Println("-------------")                                              //nolint:forbidigo
	fmt.Printf("links found:          %d\n", len(c.Results.Links))            //nolint:forbidigo
	fmt.Printf("external links found: %d\n", len(c.Results.ExternalLinks))    //nolint:forbidigo
	fmt.Printf("links that errorred:  %d\n", len(c.Results.ErroredLinks))     //nolint:forbidigo
	fmt.Printf("crawling took %.2f seconds\n", c.Results.TotalTime.Seconds()) //nolint:forbidigo
}

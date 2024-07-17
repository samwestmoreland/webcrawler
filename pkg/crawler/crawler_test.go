package crawler_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/samwestmoreland/webcrawler/pkg/crawler"
	"github.com/samwestmoreland/webcrawler/pkg/url"
	"golang.org/x/net/html"
)

// TestNewCrawler tests the creation of a new Crawler instance
func TestNewCrawler(t *testing.T) {
	c, err := crawler.NewCrawlerDiscardOutput("https://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if c.Host != "example.com" {
		t.Errorf("expected host to be %s, got %s", "example.com", c.Host)
	}

	if c.StartURL.URL != "https://example.com/" {
		t.Errorf("expected url to be %s, got %s", "https://example.com", c.StartURL.URL)
	}
}

// TestFetch tests the fetch function
func TestFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><a href="https://example.com/page1">Page1</a></body></html>`))
	}))
	defer server.Close()

	c, _ := crawler.NewCrawlerDiscardOutput(server.URL)
	urlToFetch, _ := url.ParseURLString(server.URL, "")

	doc, err := c.Fetch(urlToFetch.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if doc == nil {
		t.Fatal("expected non-nil document")
	}

	var buf bytes.Buffer
	html.Render(&buf, doc)
	if !strings.Contains(buf.String(), "Page1") {
		t.Errorf("expected fetched document to contain 'Page1', got %s", buf.String())
	}
}

// TestExtractLinks tests the link extraction from HTML document
func TestExtractLinks(t *testing.T) {
	htmlData := `<html><body>
	<a href="https://example.com/page1">Page1</a>
	<a href="https://example.com/page2">Page2</a>
	<a href="https://example.com/page3">Page3</a>
	</body></html>`

	doc, _ := html.Parse(strings.NewReader(htmlData))
	c, _ := crawler.NewCrawlerDiscardOutput("https://example.com")

	links, err := c.ExtractLinks(doc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedLinks := []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/page3",
	}

	if len(links) != len(expectedLinks) {
		t.Fatalf("expected %d links, got %d", len(expectedLinks), len(links))
	}

	for i, link := range links {
		if link != expectedLinks[i] {
			t.Errorf("expected link %s, got %s", expectedLinks[i], link)
		}
	}
}

// TestCrawl tests the Crawl function
func TestCrawl(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body>
				<a href="/page1">Page1</a>
				<a href="/page2">Page2</a>
				</body></html>`))
		case "/page1":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body>
				<a href="/">Home</a>
				<a href="/page3">Page3</a>
				</body></html>`))
		case "/page2":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body>
				<a href="/">Home</a>
				<a href="/page3">Page3</a>
				</body></html>`))
		case "/page3":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body>
				<a href="/">Home</a>
				</body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c, _ := crawler.NewCrawlerDiscardOutput(server.URL)
	err := c.Crawl()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedLinks := []string{
		server.URL + "/",
		server.URL + "/page1",
		server.URL + "/page2",
		server.URL + "/page3",
	}

	if len(c.Results.Links) != len(expectedLinks) {
		fmt.Println("Results:")
		for _, link := range c.Results.Links {
			fmt.Println(link)
		}

		fmt.Println()

		fmt.Println("Expected links:")
		for _, link := range expectedLinks {
			fmt.Println(link)
		}

		fmt.Println()

		t.Fatalf("expected %d links, got %d", len(expectedLinks), len(c.Results.Links))
	}

	for _, expected := range expectedLinks {
		found := false
		for _, result := range c.Results.Links {
			if result == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find link %s in results", expected)
		}
	}
}

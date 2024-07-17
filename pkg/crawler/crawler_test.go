package crawler_test

import (
	"bytes"
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

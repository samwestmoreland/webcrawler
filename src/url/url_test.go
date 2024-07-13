package url_test

import (
	"testing"

	"github.com/samwestmoreland/webcrawler/src/url"
)

func TestNormalise(t *testing.T) {
	testCases := []struct {
		base     string
		href     string
		expected string
	}{
		{"http://foo.com/page", "about", "http://foo.com/about"},
		{"http://foo.com/page", "http://foo.com/contact", "http://foo.com/contact"},
		{"http://foo.com/page", "about#section", "http://foo.com/about"},
		{"http://foo.com/page", "https://foo.com/about", "https://foo.com/about"},
		{"http://foo.com/page", "about?query=value#section", "http://foo.com/about?query=value"},
		{"http://foo.com/path/page", "./contact", "http://foo.com/path/contact"},
		{"http://foo.com/path/page", "/about", "http://foo.com/about"},
		{"http://foo.com/path/", "about", "http://foo.com/path/about"},
	}

	for _, testCase := range testCases {
		result, err := url.Normalise(testCase.base, testCase.href)
		if err != nil {
			t.Errorf("Got error on Normalise(%q, %q): %v", testCase.base, testCase.href, err)
		}
		if result != testCase.expected {
			t.Errorf("Got Normalise(%q, %q) = %q, want %q", testCase.base, testCase.href, result, testCase.expected)
		}
	}
}

func TestIsSameSubdomain(t *testing.T) {

	testCases := []struct {
		base     string
		href     string
		expected bool
	}{
		{"http://foo.com/page", "https://foo.com/about", true},
		{"http://foo.com/page", "http://foo.com/contact", true},
		{"http://foo.com/page", "https://foo.com/about", true},
	}

	for _, testCase := range testCases {
		result := url.IsSameSubdomain(testCase.base, testCase.href)
		if result != testCase.expected {
			t.Errorf("Got isSameSubdomain(%q, %q) = %v, want %v", testCase.base, testCase.href, result, testCase.expected)
		}
	}

}

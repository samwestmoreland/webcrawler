package url_test

import (
	"testing"

	"github.com/samwestmoreland/webcrawler/pkg/url"
)

func TestNormalise(t *testing.T) {
	testCases := []struct {
		base     string
		href     string
		expected string
	}{
		{"foo.com", "about", "https://foo.com/about"},
		{"foo.com", "/about", "https://foo.com/about"},
		{"foo.com", "/bar/baz", "https://foo.com/bar/baz"},
	}

	for _, testCase := range testCases {
		result, err := url.Normalise(testCase.base, testCase.href)
		if err != nil {
			t.Errorf("Got normalise(%q, %q) = %q, want %q", testCase.base, testCase.href, result, testCase.expected)
		}
	}
}

func TestIsSameSubdomain(t *testing.T) {
	testCases := []struct {
		base     string
		href     string
		expected bool
	}{
		{"www.foo.com", "https://foo.com/about", true},
		{"foo.com", "http://foo.com/contact/us", true},
		{"foo.com", "https://yahoo.com/about", false},
	}

	for _, testCase := range testCases {
		result := url.IsSameSubdomain(testCase.base, testCase.href)
		if result != testCase.expected {
			t.Errorf("Got isSameSubdomain(%q, %q) = %v, want %v", testCase.base, testCase.href, result, testCase.expected)
		}
	}

}

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
		{"foo.com", "about", "foo.com/about"},
		{"foo.com", "/about", "foo.com/about"},
		{"foo.com", "/bar/baz", "foo.com/bar/baz"},
	}

	for _, testCase := range testCases {
		result, _ := url.Normalise(testCase.base, testCase.href)
		if result.URL != testCase.expected {
			t.Errorf("Got normalise(%q, %q) = %q, want %q", testCase.base, testCase.href, result.URL, testCase.expected)
		}
	}
}

func TestIsSameHost(t *testing.T) {
	testCases := []struct {
		base        string
		href        string
		expected    bool
		expectedErr bool
	}{
		{"www.foo.com", "foo.com", true, false},
		{"foo.com", "www.foo.com", true, false},
		{"foo.com", "yahoo.com", false, false},
	}

	for _, testCase := range testCases {
		result, err := url.IsSameHost(testCase.base, testCase.href)
		if err != nil {
			t.Errorf("isSameHost(%q, %q) produced error: %v", testCase.base, testCase.href, err)
		}

		if result != testCase.expected {
			t.Errorf("Got isSameHost(%q, %q) = %v, want %v", testCase.base, testCase.href, result, testCase.expected)
		}
	}
}

func TestIsSameSubdomainBadInputs(t *testing.T) {
	testCases := []struct {
		base        string
		href        string
		expectedErr bool
	}{
		{"http://www.foo.com", "foo.com", true},
		{"kangaroo", "www.foo.com", true},
		{"foo.com", "yahoo.com/woohoo", true},
		{"foo.com", "foo.com", false},
		{"foocom/bar", "foo.com", true},
		{"zanzibar.xyz", "foo.com", false},
		{"foo.com", "zanzibar.xyz/a/b/c/", true},
	}

	for _, testCase := range testCases {
		_, err := url.IsSameHost(testCase.base, testCase.href)
		if (err != nil) != testCase.expectedErr {
			t.Errorf("Got isSameHost(%q, %q) = %v, want %v", testCase.base, testCase.href, err, testCase.expectedErr)
		}
	}
}

func TestParseURLString(t *testing.T) {
	testCases := []struct {
		input    string
		expected *url.URL
	}{
		{"https://www.foo.com", &url.URL{Scheme: "https", Host: "www.foo.com", Path: "/", URL: "https://www.foo.com/"}},
		{"https://www.foo.com/", &url.URL{Scheme: "https", Host: "www.foo.com", Path: "/", URL: "https://www.foo.com/"}},
		{"https://www.foo.com/bar", &url.URL{Scheme: "https", Host: "www.foo.com", Path: "/bar", URL: "https://www.foo.com/bar"}},
		{"https://www.foo.com/bar/", &url.URL{Scheme: "https", Host: "www.foo.com", Path: "/bar/", URL: "https://www.foo.com/bar/"}},
		{"https://www.foo.com/bar/baz", &url.URL{Scheme: "https", Host: "www.foo.com", Path: "/bar/baz", URL: "https://www.foo.com/bar/baz"}},
		{"www.foo.com", &url.URL{Scheme: "https", Host: "www.foo.com", Path: "/", URL: "https://www.foo.com/"}},
	}

	for _, testCase := range testCases {
		result, err := url.ParseURLString(testCase.input, "")
		if err != nil {
			t.Errorf("Got parseURLString(%q) produced error: %v", testCase.input, err)
		}

		if result.Scheme != testCase.expected.Scheme {
			t.Errorf("For input %q, got scheme %q, want %q", testCase.input, result.Scheme, testCase.expected.Scheme)
		}

		if result.Host != testCase.expected.Host {
			t.Errorf("For input %q, got host %q, want %q", testCase.input, result.Host, testCase.expected.Host)
		}

		if result.Path != testCase.expected.Path {
			t.Errorf("For input %q, got path %q, want %q", testCase.input, result.Path, testCase.expected.Path)
		}

		if result.URL != testCase.expected.URL {
			t.Errorf("For input %q, got URL %q, want %q", testCase.input, result.URL, testCase.expected.URL)
		}
	}
}

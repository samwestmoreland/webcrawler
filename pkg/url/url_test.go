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

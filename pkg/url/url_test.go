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
		{"https://foo.com", "about", "https://foo.com/about"},
		{"http://foo.com", "http://foo.com/contact", "http://foo.com/contact"},
		{"https://foo.com", "https://foo.com/about", "https://foo.com/about"},
		{"https://foo.com", "https://foo.com/about#bar", "https://foo.com/about"},
		{"https://monzo.com/legal/terms-and-conditions/", "https://monzo.com/legal/fscs-information/", "https://monzo.com/legal/fscs-information/"},
		// https://monzo.com/legal/terms-and-conditions/
		// https://monzo.com/legal/fscs-information/
		// https://monzo.com/legal/privacy-notice/

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

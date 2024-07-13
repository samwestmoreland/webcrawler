package url

import (
	"fmt"
	"net/url"
)

func Parse(url string) (*url.URL, error) {
	return url.Parse(url)
}

// Normalise resolves relative URLs into absolute URLs, remove fragments and ensure consistency.
func Normalise(base, href string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	if baseURL.Scheme == "" {
		return "", fmt.Errorf("base URL must have a scheme")
	}

	hrefURL, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	// Resolve the relative URL against the base URL
	resolvedURL := baseURL.ResolveReference(hrefURL)

	// Strip out the fragment part, if any
	resolvedURL.Fragment = ""

	// Prefix with the base URL IF the resolved URL is relative
	if !resolvedURL.IsAbs() {
		resolvedURL.Scheme = baseURL.Scheme
		resolvedURL.Host = baseURL.Host
	}

	return resolvedURL.String(), nil
}

func IsSameSubdomain(base, href string) bool {
	baseURL, err := url.Parse(base)
	if err != nil {
		return false
	}
	hrefURL, err := url.Parse(href)
	if err != nil {
		return false
	}

	return baseURL.Hostname() == hrefURL.Hostname()
}

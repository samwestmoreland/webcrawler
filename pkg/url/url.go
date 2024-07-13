package url

import (
	"fmt"
	"net/url"
)

type URL struct {
	Subdomain string
	Path      string
}

func IsValidURL(u string) bool {
	_, err := url.Parse(u)
	return err == nil
}

func Parse(u string) (*URL, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	return &URL{
		Subdomain: parsed.Hostname(),
		Path:      parsed.Path,
	}, nil
}

// Normalise resolves relative URLs into absolute URLs, removes fragments and
// ensures consistency
func Normalise(base, href string) (*URL, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	if baseURL.Scheme == "" {
		return nil, fmt.Errorf("base URL must have a scheme")
	}

	hrefURL, err := url.Parse(href)
	if err != nil {
		return nil, err
	}

	// Resolve the relative URL against the base URL
	resolvedURL := baseURL.ResolveReference(hrefURL)

	ret := &URL{
		Subdomain: resolvedURL.Hostname(),
		Path:      resolvedURL.Path,
	}

	return ret, nil
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

package url

import (
	"net/url"
	"path"
	"strings"
)

const (
	wwwPrefix = "www."
	scheme    = "https://"
)

type URL struct {
	// URL must be a valid URL, i.e. with a scheme and subdomain
	URL       string
	Subdomain string
	Path      string
}

func IsValidURL(u string) bool {
	_, err := url.Parse(u)
	return err == nil
}

func ParseURLString(u string) (*URL, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	return &URL{
		URL:       parsed.String(),
		Subdomain: parsed.Hostname(),
		Path:      parsed.Path,
	}, nil
}

// Normalise resolves relative URLs into absolute URLs (against the given base).
// The base is expected to be just a subdomain, e.g. "monzo.com"
func Normalise(subdomain, href string) (*URL, error) {
	hrefURL, err := url.Parse(href)
	if err != nil {
		return nil, err
	}

	// If hrefURL is already absolute, we just return it as is.
	if hrefURL.IsAbs() {
		return &URL{
			URL:       hrefURL.String(),
			Subdomain: hrefURL.Hostname(),
			Path:      hrefURL.Path,
		}, nil
	}

	// If the href is relative, resolve it against the subdomain.
	// We treat the subdomain as the base for relative resolution.
	baseURL := &url.URL{
		Scheme: "http",
		Host:   subdomain,
	}

	// Resolve the relative URL against the base URL.
	resolvedURL := path.Join(baseURL.Hostname(), hrefURL.Path)

	ret := &URL{
		URL:       resolvedURL,
		Subdomain: baseURL.Host,
		Path:      hrefURL.Path,
	}

	return ret, nil
}

func IsSameSubdomain(subdomainA, subdomainB string) bool {
	return strings.TrimPrefix(subdomainA, wwwPrefix) ==
		strings.TrimPrefix(subdomainB, wwwPrefix)
}

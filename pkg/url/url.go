package url

import (
	"net/url"
	"path"
	"strings"
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
	resolvedPath := path.Join(baseURL.Path, hrefURL.Path)
	if strings.HasPrefix(href, "/") {
		resolvedPath = hrefURL.Path
	}

	ret := &URL{
		Subdomain: baseURL.Host,
		Path:      resolvedPath,
	}

	return ret, nil
}

func IsSameSubdomain(subdomain, href string) bool {
	hrefURL, err := url.Parse(href)
	if err != nil {
		return false
	}

	return subdomain == hrefURL.Hostname()
}

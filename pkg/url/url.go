package url

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const (
	wwwPrefix     = "www."
	defaultScheme = "https"
)

var hostnameRegex = regexp.MustCompile(`^(www\.)?[a-zA-Z0-9-]+(\.[a-zA-Z0-9]+)+$`)

type URL struct {
	// URL must be a valid URL, i.e. with a scheme and subdomain
	URL    string
	Scheme string
	Host   string
	Path   string
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

	if parsed.Scheme == "" {
		parsed.Scheme = defaultScheme
	}

	// Re-parse the URL with the default scheme, otherwise we end up with no host
	parsed, err = url.Parse(parsed.String())
	if err != nil {
		return nil, err
	}

	// For consistency, we'll use "/" as the default path if none is provided
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	return &URL{
		URL:    parsed.String(),
		Scheme: parsed.Scheme,
		Host:   parsed.Hostname(),
		Path:   parsed.Path,
	}, nil
}

// Normalise resolves relative URLs into absolute URLs (against the given base).
// The base is expected to be just a subdomain, e.g. "foo.com"
func Normalise(subdomain, href string) (*URL, error) {
	hrefURL, err := url.Parse(href)
	if err != nil {
		return nil, err
	}

	// If hrefURL is already absolute, we just return it as is.
	if hrefURL.IsAbs() {
		return &URL{
			URL:  hrefURL.String(),
			Host: hrefURL.Hostname(),
			Path: hrefURL.Path,
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
		URL:  resolvedURL,
		Host: baseURL.Host,
		Path: hrefURL.Path,
	}

	return ret, nil
}

func IsSameHost(hostA, hostB string) (bool, error) {
	if !hostnameRegex.MatchString(hostA) {
		return false, fmt.Errorf("invalid host: %q", hostA)
	}

	if !hostnameRegex.MatchString(hostB) {
		return false, fmt.Errorf("invalid host: %q", hostB)
	}

	return strings.TrimPrefix(hostA, wwwPrefix) ==
		strings.TrimPrefix(hostB, wwwPrefix), nil
}

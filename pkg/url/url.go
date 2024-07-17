package url

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const (
	wwwPrefix = "www."
)

var hostnameRegex = regexp.MustCompile(`^(www\.)?[a-zA-Z0-9-]+(\.[a-zA-Z0-9]+)+$`)

// The URL type is a simplified version of the net/url URL type with only the fields we need
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

// ParseURLString parses a string into a URL type. It takes as an argument the
// scheme to use if the URL doesn't have one
func ParseURLString(u string, scheme string) (*URL, error) {
	if scheme == "" {
		scheme = "https"
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme == "" {
		parsed.Scheme = scheme
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

// ResolvePath resolves relative URLs into absolute URLs (against the given base).
// The base is expected to be just a subdomain, e.g. "foo.com"
func ResolvePath(subdomain, href string) (*URL, error) {
	// Trim any leading white spaces
	href = strings.TrimSpace(href)

	hrefURL, err := url.Parse(href)
	if err != nil {
		return nil, err
	}

	// If hrefURL is already absolute, we just return it as is
	if hrefURL.IsAbs() {
		return &URL{
			Scheme: hrefURL.Scheme,
			URL:    hrefURL.String(),
			Host:   hrefURL.Hostname(),
			Path:   hrefURL.Path,
		}, nil
	}

	// If the href is relative, resolve it against the subdomain.
	baseURL := &url.URL{
		Scheme: "http",
		Host:   subdomain,
	}

	resolvedURL := path.Join(baseURL.Hostname(), hrefURL.Path)

	ret := &URL{
		URL:    resolvedURL,
		Scheme: baseURL.Scheme,
		Host:   baseURL.Host,
		Path:   hrefURL.Path,
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

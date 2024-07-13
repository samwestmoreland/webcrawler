package url

import "net/url"

// Normalise resolves relative URLs into absolute URLs, remove fragments and ensure consistency.
func Normalise(base, href string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	hrefURL, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	// Resolve the relative URL against the base URL
	resolvedURL := baseURL.ResolveReference(hrefURL)

	// Strip out the fragment part, if any
	resolvedURL.Fragment = ""

	return resolvedURL.String(), nil
}

func IsSameSubdomain(base string, link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.Hostname() == base
}

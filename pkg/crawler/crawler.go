package crawler

type Crawler struct {
	// this would be monzo.com or community.monzo.com for example
	subdomain string
}

func NewCrawler(url string) Crawler {
	return Crawler{subdomain: subdomain}
}

func (c Crawler) Crawl() {
	// TODO
}

package main

import (
	"log"
)

func main() {
	startURL := "https://www.monzo.com"

	crawler, err := NewCrawler(startURL)
	if err != nil {
		log.Fatal(err)
	}

	crawler.Crawl()
}

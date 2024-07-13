package main

import (
	"log"

	"github.com/samwestmoreland/webcrawler/pkg/crawler"
)

func main() {
	startURL := "https://www.monzo.com"

	myCrawler, err := crawler.NewCrawler(startURL)
	if err != nil {
		log.Fatal(err)
	}

	myCrawler.Crawl()
}

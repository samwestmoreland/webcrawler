package main

import (
	"log"

	"github.com/samwestmoreland/webcrawler/pkg/crawler"
)

func main() {
	startURL := "https://www.monzo.com"

	myCrawler, err := crawler.NewDefaultCrawler(startURL)
	if err != nil {
		log.Fatal(err)
	}

	if err = myCrawler.Crawl(); err != nil {
		log.Fatal(err)
	}

	myCrawler.OutputResults()
}

package main

import (
	"log"

	"github.com/samwestmoreland/webcrawler/pkg/crawler"
)

func main() {
	startURL := "https://www.thoughtmachine.net"

	myCrawler, err := crawler.NewCrawler(startURL)
	if err != nil {
		log.Fatal(err)
	}

	if err = myCrawler.Crawl(); err != nil {
		log.Fatal(err)
	}

	myCrawler.OutputResults()
}

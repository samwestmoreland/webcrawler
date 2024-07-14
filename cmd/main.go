package main

import (
	"log"
	"os"

	"github.com/samwestmoreland/webcrawler/pkg/crawler"
)

func main() {
	startURL := "https://www.thoughtmachine.net"

	logFile, err := os.OpenFile("crawler.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	myCrawler, err := crawler.NewCrawlerWithLogFile(startURL, logFile)
	if err != nil {
		log.Fatal(err)
	}

	if err = myCrawler.Crawl(); err != nil {
		log.Fatal(err)
	}

	myCrawler.OutputResults()
}

package main

import (
	"log"
	"os"

	"github.com/samwestmoreland/webcrawler/pkg/crawler"
)

const defaultLogFileName = "crawler.log"

func main() {
	startURL := "https://www.monzo.com"

	logFile, err := os.OpenFile(defaultLogFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(logFile, "", log.Ldate|log.Ltime)

	myCrawler, err := crawler.NewDefaultCrawler(startURL, logger)
	if err != nil {
		logger.Fatal(err)
	}

	if err = myCrawler.Crawl(); err != nil {
		logger.Fatal(err)
	}

	myCrawler.OutputResults()
}

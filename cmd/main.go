package main

import (
	"flag"
	"log"
	"os"

	"github.com/samwestmoreland/webcrawler/pkg/crawler"
)

func main() {
	startURL := flag.String("url", "", "URL to crawl.")
	logFileName := flag.String("log", "crawler.log", "Log file name. Default: crawler.log")
	outputFileName := flag.String("output", "crawler.out", "Output file name. Default: crawler.out")

	flag.Parse()

	if *startURL == "" {
		log.Fatal("A URL must be specified")
	}

	logFile, err := os.OpenFile(*logFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(logFile, "", log.Ldate|log.Ltime)

	resultsFile, err := os.OpenFile(*outputFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		log.Fatal(err)
	}

	myCrawler, err := crawler.NewCrawler(
		*startURL,
		logger,
		resultsFile,
	)
	if err != nil {
		logger.Fatal(err)
	}

	if err = myCrawler.Crawl(); err != nil {
		logger.Fatal(err)
	}

	myCrawler.OutputResults()
}

package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	url := "https://example.com"

	log.Println("Visiting", url)

	body, err := fetch(url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))

	// parse html
	nodeTree, err := html.Parse(body)
	if err != nil {
		log.Fatal(err)
	}

	for node := nodeTree; node != nil; node = node.NextSibling {
		if node.Type == html.TextNode && node.Data == "a" {
			fmt.Println(node.Data)
		}
	}
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

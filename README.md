# Usage
To run the crawler:
`go run cmd/main.go -url <url>`

For help on what flags the crawler accepts:
`go run cmd/main.go -help`

To run unit tests:
`go test ./...`

# How it works:
The crawler performs a breadth-first search of all the pages it finds for the given subdomain. Beginning with a single URL, the crawler visits the URL and extracts all the links from it. It then adds the links to a queue. While the queue is not empty, the crawler visits each link in the queue in turn, adding any more links it finds to the queue.

# Improvements to be made:
- Improve performance by introducing concurrency into the BFS.
- Keep track of link parents, so that if a dead link is found we can know where it was linked from, for example.

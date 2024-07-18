# Usage
To run the crawler:
`go run cmd/main.go -url <url>`

For help on what flags the crawler accepts:
`go run cmd/main.go -help`

To run unit tests:
`go test ./...`

By default, the crawler will log to `crawler.log` in the current directory, and output the results of the crawl to `crawler.out`. Both of these can be set via CLI flags (see CLI help).

# How it works:
## Summary
The crawler performs a breadth-first search of all the pages it finds for the given subdomain. Beginning with a single URL, the crawler visits the URL and extracts all the links from it. It then adds the links to a queue. While the queue is not empty, the crawler visits each link in the queue in turn, adding any more links it finds to the queue.

## Package structure
The project has two subpackages: `crawler` and `url`. The `crawler` package contains the logic for the crawler. It defines a Crawler type which contains some request-related configuration: request timeouts, number of retries etc., as well as some output configuration, namely the log file to log to and the results file which is an io.Writer, so the caller could set this to os.Stdout if they so chose.

The `url` package contains some helper methods for working with URLs. It defines a URL type which is just contains a subset of useful fields of the URL type in the standard library.

## If a 202 response is received
If a 202 "Status Accepted" response is received, the crawler will poll the URL for a maximum of 5 times with a 5 second delay between each request. If a 200 response is never received, the URL will be marked as errored.

# Planned improvements
- Explore possible performance gains from introducing concurrency into the BFS. Have tried this but preliminary results were slower than the single-threaded version, likely due to contention over the `visitedSet` map.
- Keep track of link parents, so that if a dead link is found we can know where it was linked from, for example.

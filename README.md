To run the crawler:
`go run cmd/main.go`

To run unit tests:
`go test ./...`

How it works:
The crawler 

To-dos:
- Make sure every function has a comment
- Get rid of TODOs before submission
- Expose subset of config options
- Split out crawl logic
- More test coverage

Improvements:
- Keep track of link parents, so that if a dead link is found we can know where it was linked from, for example.

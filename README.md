To run the crawler:
`go run cmd/main.go -url <url>`

For help on what flags the crawler accepts:
`go run cmd/main.go -help`

To run unit tests:
`go test ./...`

How it works:
The crawler 

To-dos:
- Make sure every function has a comment
- Get rid of TODOs before submission
- Expose subset of config options
- More test coverage

Improvements:
- Keep track of link parents, so that if a dead link is found we can know where it was linked from, for example.

# Contributing Katabole

We want Katabole to be a welcoming environment for all contributors, even those without much open source experience.
However, please carefully read and follow the standards below.

For those who are new, check out [our page for first-time developers](FIRST-TIME.md).

# Development and testing standards

Pull requests should:
-   Have informative summaries and comments on git commits
-   Have passing unit tests
-   Provide features or bug fixes in keeping with our design goals (see [FAQ](FAQ.md))
-   Formatted with [`go fmt`](https://pkg.go.dev/cmd/gofmt) and 
    [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports))
-   Well documented; your code teaches an unfamiliar engineer what it does and why
-   Matches the general coding style of the project
-   Clear and modularized
-   Follows guidelines as laid out in [Effective Go](https://golang.org/doc/effective_go.html)

We generally follow these style guidelines:
- [How to Write Go Code](http://golang.org/doc/code.html)
- [Godoc documentation style](http://blog.golang.org/godoc-documenting-go-code)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Test Comments](https://github.com/golang/go/wiki/TestComments)

Lastly, when you need a feature provided from a dependency, check for what katabole packages already use. For example,
for test assertions use [testify/assert](https://pkg.go.dev/github.com/stretchr/testify/assert) and
[testify/require](https://pkg.go.dev/github.com/stretchr/testify/require), and for logging,
[slog](https://pkg.go.dev/log/slog).

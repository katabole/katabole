# Katabole Frequently Asked Questions

## What are the design goals of Katabole?

Here they are, in priority order:

1. Maintainability over time
2. Productivity through transparency

I've observed many web frameworks suffer in these two areas.

In the area of maintainability, picking up cool new technologies can be very tempting. Many web developers don't see a
problem with using the latest and greatest thing. But doing so leads to churn as things keep changing or packages
receive breaking updates. Part of the wisdom of Go has been consistent backward-compatibility, and with Katabole I hope
to accomplish a similar thing by favoing technologies that have achieved stability and avoiding unnecessary
dependencies. In other words, boring technology rocks. As a [Go Proverb](https://go-proverbs.github.io/) says, "a little
copying is better than a little dependency."

As for productivity, many frameworks take DRY (Don't Repeat Yourself) to the limit. This creates lots of magic. In
theory it makes possible very fast development by an expert framework user. In practice it causes a maintenance burden
when those magic parts must change. More importantly, it confuses new developers and hides things they really should
learn. I like to think of Katabole as a learning framework; though you can get going with app generation, over time
you'll learn what each part does. This is good.

## Why pick technology ___?

### Languages/Frameworks/Packages

#### Go

The community behind Go itself understands the design goals above. Go is not only stable, but easy to pick up, maintain,
and monitor in production environments.

#### Chi

Chi has a great API that hasn't changed much. More importantly, it's compatible with the Go standard library's [http
handler signature](https://pkg.go.dev/net/http#Handle), meaning even Chi could be replaced with something else. I did
try using just the [built-in ServeMux](https://pkg.go.dev/net/http#ServeMux) but found it too limiting. Routing is a
pretty key feature needed for web and API applications.

#### Vite

Vite is modern and, more importantly, seems to be here to stay.

#### Docker

Docker is ubiquitous.

#### Postgres

Postgres is also ubiquitous, and solid. There's a reason higher-scale databases like CockroachDB and Aurora aim to
support it first.

### Tools

Somewhat by accident I ended up on tools that themselves are written in Go. I say "somewhat" because it wasn't an
up-front restriction, but given the culture of Go, it makes sense that they satisfy my design goals. They're also fast
and easy to install as single binaries.

#### Task

Task is a great modern generalized task runner, which also satisfied parallelism needs.

#### Atlas

Atlas doesn't have the longevity of some other SQL management tools, so it was a difficult choice. But the ability to
manage one well-documented schema.sql file was a big win. The workflow of editing it and generating migrations for
deployment (or just using it [declaratively](https://atlasgo.io/declarative/apply)) was very smooth.

Note that I have no intention of using its HCL syntax, just plain SQL.

For projects too limited by features available in the free version, like views and functions, try
[Goose](https://github.com/pressly/goose).

## Why the name?

Katabole is a Greek word referring to laying down a foundation, as in "foundation of the world". It was unique and had a
convenient two-character short form "KB" to use as a prefix.

## When I run `katabole gen`, what does it do?

`katabole gen` generates a project for you. By default it clones the [kbexample](https://github.com/katabole/kbexample)
repository, using it as a template by replacing all instances of "KBExample" with your project name.

## Where do you manage Katabole issues?

You can create issues in the relevant repository, but we have a central project board
[here](https://github.com/orgs/katabole/projects/1).

## What inspired the project?

I needed a framework for quickly building web applications. It needed to handle using a SQL database, be approachable
for less-experienced developers, and satisfy the design goals noted above. I didn't find anything that satisfied these
requirements.

In the Go community the phrase "just use the standard library" is used often, and there's merit to that. But I think
it's worth appreciating how few software developers are skilled enough to build a production web application that way.
In my opinion [Buffalo](https://github.com/gobuffalo/buffalo) was the best attempt at this, but made itself a little too
difficult to maintain with the design ideas it imported from frameworks like Ruby on Rails. I'll be proud if Katabole
can make web development as approachable as Buffalo did.

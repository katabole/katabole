<p align="center" dir="auto">
    <img src="images/katabole-logo.jpg" alt="Katabole" width="200" style="border-radius: 45px;">
</p>

# Katabole (ka-ta-buh-LAY) - A productive, timeless Go web framework.

Katabole offers a foundation generator for building your Go web application. `katabole gen` creates a project for you
from the [kbexample](https://github.com/katabole/kbexample) template, which puts together the following:
- [Chi](https://go-chi.io/#/README) for routing
- [Webpack](https://webpack.js.org/) for front-end assets
- [Atlas](https://atlasgo.io/) for SQL DB management
- [Task](https://taskfile.dev) for task management
- [Docker](https://www.docker.com/get-started/) and [Air](https://github.com/air-verse/air) for smooth development
- Packages in [katabole](https://github.com/katabole) and [gorilla](github.com/gorilla), plus a few [others](https://github.com/katabole/kbexample/blob/main/go.mod)

Katabole is unique because it's hardly a framework. It provides the generator, some well-defined stable packages, and
conventions. As a developer you learn how it fits together; it's transparent, not magic.

It's productive because in a matter of minutes, you're up and running. It's timeless because the code is yours. Katabole
may upgrade or swap out a thing or two (rarely, I hope), but your code will continue to build.

## Quick Start

Prerequisites: [Go](https://go.dev/doc/install), [Node](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm), [Task](https://taskfile.dev/installation/), and [Docker](https://docs.docker.com/get-docker/).

```bash
go install github.com/katabole/katabole
katabole gen --import-path github.com/myuser/myapp --title-name MyApp
cd myapp
task setup

# Run a dev server, then check out http://localhost:3000
task dev

# See what else you can do built-in
task -l
```

## Contributing

Development happens primarily on [kbexample](https://github.com/katabole/kbexample), the canonical Katabole app that
also acts as a template.

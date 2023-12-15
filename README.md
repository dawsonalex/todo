# golang-cli

![example workflow](https://github.com/dawsonalex/golang-cli/workflows/Build/badge.svg)

A template for CLI applications in Golang.

## Why?

Often I have the idea to create a new small CLI application in Go. I want that application to have the same niceties
that other applications have, in terms of CI and automatic releases but often the cost set up outweighs the payoff.

This template not only provides a starting structure for CLI applications, but also a basic GitHub Action that will
manage build, testing, and releases automatically based on commit message.

## Usage

This template assumes it will be used as a [Go module](https://golang.org/ref/mod). To begin, change the top line
of `go.mod` to match the name you would use when running `go mod init`.

You can then use `make release` to build the application, which outputs the executable to the `bin` dir. The `bin` dir
is under `.gitignore` so binaries will not be tracked by git.

To remove binaries run `make clean`.

Other than the default layout for the `main` package, this template doesn't impose any restrictions in the user. My
preference when creating an application is to organise code [by layers](https://www.gobeyond.dev/packages-as-layers/),
meaning I'll probably place _business domain_ code in the root of the package, and put other packages that provide other
features not directly related to the business domain in sub-packages. For example:

```text
golang-cli/
|  README.md
|  domainpart1.go
|  domainpart2.go
|
└-cmd/
|  main.go
|
└-http
  |  http.go  
```

The above has the benefit of allowing the module to expose business logic as a library would, as well as having an
entrypoint for `main` under the `cmd` directory. YMMV of course.

## GitHub Workflow

The GitHub workflow defined in `base.yml` attempts to do some common things in a simple way. Currently, it does the
following steps under a single job called `Build`:

    - Set up Go environment.
    - Run Go Tests (ignoring if there are none).
    - Runs `make release` to create all binaries.
    - Bump the version based on the commit message.
        - Use `#major`, `#minor`, or `#patch` in your commit message to bump the version and create a new release.
        - Leaving out the above tags will not create a new tag or release version.
    - Generate release logs from the commits between this tag and the last.
    - Create a GitHub release and upload the content of `bin`.
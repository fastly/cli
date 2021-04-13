## Development

The Fastly CLI requires [Go 1.16 or above](https://golang.org). Clone this repo
to any path and type `make` to run all of the tests and generate a development
build locally.

```sh
git clone git@github.com:fastly/cli
cd cli
make
./fastly version
```

The `make` task requires the following executables to exist in your `$PATH`:

- [golint](https://github.com/golang/lint)
- [gosec](https://github.com/securego/gosec)
- [staticcheck](https://staticcheck.io/)

If you have none of them installed, or don't mind them being upgraded automatically, you can run `make dependencies` to install them.

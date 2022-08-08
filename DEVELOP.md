## Development

The Fastly CLI requires [Go 1.18 or above](https://golang.org). Clone this repo
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

### New API Command Scaffolding

There are two ways to scaffold a new command, that is intended to be a non-composite command (i.e. a straight 1-1 mapping to an underlying Fastly API endpoint):

1. `make scaffold`: e.g. `fastly foo <create|delete|describe|update>`
2. `make scaffold-category`: e.g. `fastly foo bar <create|delete|describe|update>`

The latter `Makefile` target is for commands we want to group under a common category (in the above example `foo` is the category and `bar` is the command). A real example of a category command would be `fastly logging`, where `logging` is the category and within that category are multiple logging provider commands (e.g. `fastly logging splunk`, where `splunk` is the command).

The `logging` category is an otherwise non-functional command (i.e. if you execute `fastly logging`, then all you see is help output describing the available commands under the logging category).

**Makefile target structure:**

```bash
CLI_PACKAGE=... CLI_COMMAND=... CLI_API=... make scaffold
CLI_CATEGORY=... CLI_CATEGORY_COMMAND=... CLI_PACKAGE=... CLI_COMMAND=... CLI_API=... make scaffold-category
```

**Example usage:**

Imagine you want to add the following top-level command `fastly foo-bar` with CRUD commands beneath it (e.g. `fastly foo <create|delete|describe|list|update>`), then you would execute:

```bash
CLI_PACKAGE=foobar CLI_COMMAND=foo-bar CLI_API=Bar make scaffold
```

> **NOTE**: Go package names shouldn't have special characters, hence the difference between `foobar` and the command `foo-bar`. Also, the `CLI_API` value will be interpolated into CRUD verbs (`CreateBar`, `DeleteBar` etc), along with their inputs (`fastly.CreateBarInput`, `fastly.DeleteBarInput` etc).

Now imagine you want to add a new subcommand to an existing category such as 'logging' `fastly logging foo-bar` with CRUD commands beneath it (e.g. `fastly logging foo-bar <create|delete|describe|list|update>`), then you would execute a similar command but you would change to the `scaffold-category` target and also prefix two additional inputs:

```bash
CLI_CATEGORY=logging CLI_CATEGORY_COMMAND=logging CLI_PACKAGE=foobar CLI_COMMAND=foo-bar CLI_API=Bar make scaffold-category
```

> **NOTE**: Within the generated files, keep an eye out for any `<...>` references that need to be manually updated.

### config.toml

When compiling the CLI for a new release, it will pull the required configuration data from the following API endpoint:

```
https://developer.fastly.com/api/internal/cli-config
```

The config served from that endpoint is the result of an internal Fastly build process that uses the `.fastly/config.toml` file in the CLI repo as a template, and then dynamically adds in any available Starter Kits.

The configuration is then saved to disk (at the following location) and embedded into the CLI when compiled.

```
./cmd/fastly/static/config.toml
```

## Development

Building the Fastly CLI requires [Go](https://golang.org) (version
1.18 or later), and [Rust](https://www.rust-lang.org/). Clone this
repo to any path and type `make` to run all of the tests and generate
a development build locally.

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

### `.fastly/config.toml`

The CLI dynamically generates the `./pkg/config/config.toml` within the CI release process so it can be embedded into the CLI binary.

The file is added to `.gitignore` to avoid it being added to the git repository.

When compiling the CLI for a new release, it will execute [`./scripts/config.sh`](./scripts/config.sh). The script uses [`./.fastly/config.toml`](./.fastly/config.toml) as a template file to then dynamically inject a list of starter kits (pulling their data from their public repositories).

The resulting configuration is then saved to disk at `./pkg/config/config.toml` and embedded into the CLI when compiled.

When a user installs the CLI for the first time, they'll have no existing config and so the embedded config will be used. In the future, when the user updates their CLI, the existing config they have will be used.

If the config has changed in any way, then you (the CLI developer) should ensure the `config_version` number is bumped before publishing a new CLI release. This is because when the user updates to that new CLI version and the invoke the CLI, the CLI will identify a mismatch between the user's local config version and the embedded config version. This will cause the embedded config to be merged with the local config and consequently the user's config will be updated to include the new fields.

> **NOTE:** The CLI does provide a `fastly config --reset` option that resets the config to a version compatible with the user's current CLI version. This is fallback for users who run into issues for whatever reason.

### Running Compute commands locally

If you need to test the Fastly CLI locally while developing a Compute feature, then use the `--dir` flag (exposed on `compute build`, `compute deploy`, `compute serve` and `compute publish`) to ensure the CLI doesn't attempt to treat the repository directory as your project directory.

```shell
go run cmd/fastly/main.go compute deploy --verbose --dir ../../test-projects/testing-fastly-cli
```

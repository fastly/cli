<div align="center">
  <h3 align="center">Fastly CLI</h3>
  <p align="center">A CLI for interacting with the Fastly platform.</p>
  <p align="center">
      <a href="https://developer.fastly.com/reference/cli/"><img alt="Documentation" src="http://img.shields.io/badge/go-documentation-blue.svg"></a>
      <a href="https://github.com/fastly/cli/releases/latest"><img alt="Latest release" src="https://img.shields.io/github/v/release/fastly/cli" /></a>
      <a href="#License"><img alt="Apache 2.0 License" src="https://img.shields.io/github/license/fastly/cli" /></a>
      <a href="https://goreportcard.com/report/github.com/fastly/cli"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/fastly/cli" /></a>
  </p>
</div>

## Quick links
- [Installation](INSTALL.md)
- [Usage](#Usage)
- [Bash/ZSH completion](#bashzsh-shell-completion)
- [Development](#Development)
- [Testing](#Testing)
- [Issues](#Issues)

## Usage

The Fastly CLI interacts with [the Fastly API][api] via an [API token][tokens].
You'll need to [create an API token][create] for yourself, and then provide it
to the Fastly CLI in one of three ways:

1. Stored in a config file by running `fastly configure`
1. Explicitly via the `--token, -t` flag
1. Implicitly via the `FASTLY_API_TOKEN` environment variable

[api]: https://docs.fastly.com/api
[tokens]: https://docs.fastly.com/api/auth#tokens
[create]: https://docs.fastly.com/en/guides/using-api-tokens#creating-api-tokens

To see an overview of all commands, simply run `fastly` with no arguments.
Succinct help about any command or subcommand is available via the `-h, --help`
flag. Verbose help about any command or subcommand is available via the help
argument, e.g. `fastly help service`.

## Bash/ZSH shell completion
The CLI can generate completions for all commands, subcommands and flags.

By specifying `--completion-bash` as the first argument, the CLI will show possible subcommands. By ending your argv with `--`, hints for flags will be shown.

### Configuring your shell
To install the completions source them in your `bash_profile` (or equivalent):
```
eval "$(fastly --completion-script-bash)"
```

Or for ZSH in your `zshrc`:
```
eval "$(fastly --completion-script-zsh)"
```

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

## Testing

To run the test suite:

```sh
make test
```

Note that by default the tests are run using `go test` with the following configuration:

```
-race ./{cmd,pkg}/...
```

To run a specific test use the `-run` flag (exposed by `go test`) and also provide the path to the directory where the test files reside (replace `...` and `<path>` with appropriate values):

```sh
make test TESTARGS="-run <...> <path>"
```

**Example**:

```sh
make test TESTARGS="-run TestBackendCreate ./pkg/backend/..."
```

Some integration tests aren't run outside of the CI environment, to enable these tests locally you'll need to set a specific environment variable relevant to the test.

The available environment variables are:

- `TEST_COMPUTE_INIT`: runs `TestInit`.
- `TEST_COMPUTE_BUILD`: runs `TestBuildRust` and `TestBuildAssemblyScript`.
- `TEST_COMPUTE_BUILD_RUST`: runs `TestBuildRust`.
- `TEST_COMPUTE_BUILD_ASSEMBLYSCRIPT`: runs `TestBuildAssemblyScript`.

**Example**:

```sh
TEST_COMPUTE_BUILD_RUST=1 make test TESTARGS="-run TestBuildRust/fastly_crate_prerelease ./pkg/compute/..." 
```

When running the tests locally, if you don't have the relevant language ecosystems set-up properly then the tests will fail to run and you'll need to review the code to see what the remediation steps are, as that output doesn't get shown when running the test suite.

> **NOTE**: you might notice a discrepancy between CI and your local environment which is caused by the difference in Rust toolchain versions as defined in .github/workflows/pr_test.yml which specifies the version required to be tested for in CI. Running `rustup toolchain install <version>` and `rustup target add wasm32-wasi --toolchain <version>` will resolve any failing integration tests you may be running locally.

## Contributing

Refer to [CONTRIBUTING.md](./CONTRIBUTING.md)

## Issues

If you encounter any non-security-related bug or unexpected behavior, please [file an issue][bug]
using the bug report template.

[bug]: https://github.com/fastly/cli/issues/new?labels=bug&template=bug_report.md

### Security issues

Please see our [SECURITY.md](SECURITY.md) for guidance on reporting security-related issues.

## License

[Apache 2.0](LICENSE).

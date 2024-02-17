# Development

Building the Fastly CLI requires a [Go](https://golang.org) version compatible
with the [go.mod](./go.mod) version, and [Rust](https://www.rust-lang.org/).
Clone this repo to any path and type `make` to run all of the tests and generate
a development build locally.

```shell
git clone git@github.com:fastly/cli 
cd cli 
make 
./fastly version 
```

The `make` task requires the following executables to exist in your `$PATH`:

- [golint](https://github.com/golang/lint)
- [gosec](https://github.com/securego/gosec)
- [staticcheck](https://staticcheck.io/)

If you have none of them installed, or don't mind them being upgraded
automatically, you can run `make dependencies` to install them.

## Fastly Configuration

The CLI dynamically generates the `./pkg/config/config.toml` within the CI
release process so it can be embedded into the CLI binary.

The file is added to `.gitignore` to avoid it being added to the git repository.

When compiling the CLI for a new release, it will trigger
[`./scripts/config.sh`](./scripts/config.sh). The config script uses
[`./.fastly/config.toml`](./.fastly/config.toml) as a template file to then
dynamically inject a list of starter kits (pulling their data from their public
repositories).

The resulting configuration is then saved to disk at `./pkg/config/config.toml`
and embedded into the CLI when compiled.

When a user installs the CLI for the first time, they'll have no existing config
and so the embedded config will be used. In the future, when the user updates
their CLI, the existing config they have will be used.

If the config has changed in any way, then you (the CLI developer) should ensure
the `config_version` number is bumped before publishing a new CLI release. This
is because when the user updates to that new CLI version and the invoke the CLI,
the CLI will identify a mismatch between the user's local config version and the
embedded config version. This will cause the embedded config to be merged with
the local config and consequently the user's config will be updated to include
the new fields.

> **NOTE:** The CLI does provide a `fastly config --reset` option that resets
> the config to a version compatible with the user's current CLI version. This
> is fallback for users who run into issues for whatever reason.

## Running Compute commands locally

If you need to test the Fastly CLI locally while developing a Compute feature,
then use the `--dir` flag (exposed on `compute build`, `compute deploy`,
`compute serve` and `compute publish`) to ensure the CLI doesn't attempt to
treat the repository directory as your project directory.

```shell
go run cmd/fastly/main.go compute deploy --verbose --dir ../../test-projects/testing-fastly-cli
```

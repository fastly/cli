# Fastly CLI

A CLI for interacting with the Fastly platform.

## Quick links
- [Installation](#Installation)
- [Usage](#Usage)
- [Development](#Development)
- [Issues](#Issues)

## Installation

### macOS
#### Homebrew

Install: `brew install fastly/tap/fastly`

Upgrade: `brew upgrade fastly`

### Windows
#### scoop
Install:

```
scoop bucket add fastly-cli https://github.com/fastly/scoop-cli.git
scoop install fastly
```
Upgrade: `scoop update fastly`

### Linux
#### Debian/Ubuntu Linux

Install and upgrade:

1. Download the `.deb` file from the [releases page][releases]
2. `sudo apt install ./fastly_*_linux_amd64.deb` install the downloaded file

#### Fedora Linux

Install and upgrade:

1. Download the `.rpm` file from the [releases page][releases]
2. `sudo dnf install fastly_*_linux_amd64.rpm` install the downloaded file

#### Centos Linux

Install and upgrade:

1. Download the `.rpm` file from the [releases page][releases]
2. `sudo yum localinstall fastly_*_linux_amd64.rpm` install the downloaded file

#### openSUSE/SUSE Linux

Install and upgrade:

1. Download the `.rpm` file from the [releases page][releases]
2. `sudo zypper in fastly_*_linux_amd64.rpm` install the downloaded file

### From a prebuilt binary
[Download the latest release][latest] from the [releases page][releases].
Unarchive the binary and place it in your $PATH. You can verify the integrity
of the binary using the SHA256 checksums file `fastly_x.x.x_SHA256SUMS` provided 
alongside the release.

[latest]: https://github.com/fastly/cli/releases/latest
[releases]: https://github.com/fastly/cli/releases

Verify it works by running `fastly version`.

```
$ fastly version
Fastly CLI version vX.Y.Z (abc0001)
Built with go version go1.13.1 linux/amd64
```

The Fastly CLI will notify you if a new version is available, and can update
itself via `fastly update`.

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

## Development

The Fastly CLI requires [Go 1.13 or above](https://golang.org). Clone this repo
to any path and type `make` to run all of the tests and generate a development
build locally.

```sh
git clone git@github.com:fastly/cli
cd cli
make
./fastly version
```

## Contributing

We're happy to receive feature requests and PRs. If your change is nontrivial,
please open an [issue](https://github.com/fastly/cli/issues/new) to discuss the idea and implementation strategy before
submitting a PR.

## Issues

If you encounter any non-security-related bug or unexpected behavior, please [file an issue][bug]
using the bug report template.

[bug]: https://github.com/fastly/cli/issues/new?labels=bug&template=bug_report.md

### Security issues

Please see our [SECURITY.md](SECURITY.md) for guidance on reporting security-related issues.

## License

[Apache 2.0](LICENSE).

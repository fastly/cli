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
- [Development](DEVELOP.md)
- [Testing](TESTING.md)
- [Issues](#Issues)

## Usage

The Fastly CLI interacts with [the Fastly API][api] via an [API token][authentication].  You'll need to create an API token for yourself, and then provide it to the Fastly CLI in one of three ways:

1. Stored in a config file by running `fastly configure`
1. Explicitly via the `--token, -t` flag
1. Implicitly via the `FASTLY_API_TOKEN` environment variable

[api]: https://developer.fastly.com/reference/api/
[authentication]: https://developer.fastly.com/reference/api/#authentication

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

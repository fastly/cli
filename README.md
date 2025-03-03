<div align="center">
  <h3 align="center">Fastly CLI</h3>
  <p align="center">A CLI for interacting with the Fastly platform.</p>
  <p align="center">
      <a href="https://www.fastly.com/documentation/reference/cli"><img alt="Documentation" src="https://img.shields.io/badge/cli-reference-yellow"></a>
      <a href="https://github.com/fastly/cli/releases/latest"><img alt="Latest release" src="https://img.shields.io/github/v/release/fastly/cli" /></a>
      <a href="#License"><img alt="Apache 2.0 License" src="https://img.shields.io/github/license/fastly/cli" /></a>
      <a href="https://goreportcard.com/report/github.com/fastly/cli"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/fastly/cli" /></a>
  </p>
</div>

## Quick links

- [Installation](https://www.fastly.com/documentation/reference/tools/cli#installing)
- [Shell auto-completion](https://www.fastly.com/documentation/reference/tools/cli#shell-auto-completion)
- [Configuring](https://www.fastly.com/documentation/reference/tools/cli#configuring)
- [Commands](https://www.fastly.com/documentation/reference/cli#command-groups)
- [Development](https://github.com/fastly/cli/blob/main/DEVELOPMENT.md)
- [Testing](https://github.com/fastly/cli/blob/main/TESTING.md)
- [Documentation](https://github.com/fastly/cli/blob/main/DOCUMENTATION.md)

## Versioning and Release Schedules

The maintainers of this module strive to maintain [semantic versioning
(SemVer)](https://semver.org/). This means that breaking changes
(removal of functionality, or incompatible changes to existing
functionality) will be released in a version with the first version
component (`major`) incremented. Feature additions will increment the
second version component (`minor`), and bug fixes which do not affect
compatibility will increment the third version component (`patch`).

On the second Wednesday of each month, a release will be published
including all breaking, feature, and bug-fix changes that are ready
for release. If that Wednesday should happen to be a US holiday, the
release will be delayed until the next available working day.

If critical or urgent bug fixes are ready for release in between those
primary releases, patch releases will be made as needed to make those
fixes available.

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/fastly/cli/blob/main/CONTRIBUTING.md)

## Issues

If you encounter any non-security-related bug or unexpected behavior, please [file an issue][bug]
using the bug report template.

Please also check the [CHANGELOG](https://github.com/fastly/cli/blob/main/CHANGELOG.md) for any breaking-changes or migration guidance.

### Security issues

Please see our [SECURITY.md](SECURITY.md) for guidance on reporting security-related issues.

## Binaries with unreleased changes

Binaries containing merged changes that are planned for the next release are available [here](https://github.com/fastly/cli/actions/workflows/merge_to_main.yml). 
Use at your own risk. 
Updating will revert the binary to the latest released version.

## License

[Apache 2.0](LICENSE).

[bug]: https://github.com/fastly/cli/issues/new?labels=bug&template=bug_report.md

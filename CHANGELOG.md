# Changelog

## [v0.27.0](https://github.com/fastly/cli/releases/tag/v0.27.0) (2021-04-15)

[Full Changelog](https://github.com/fastly/cli/compare/v0.26.3...v0.27.0)

**Enhancements:**

- Support IAM role in Kinesis logging endpoint [\#255](https://github.com/fastly/cli/pull/255)
- Support IAM role in S3 and Kinesis logging endpoints [\#253](https://github.com/fastly/cli/pull/253)
- Add support for `file_max_bytes` configuration for Azure logging endpoint [\#251](https://github.com/fastly/cli/pull/251)
- Warn on empty directory [\#247](https://github.com/fastly/cli/pull/247)
- Add `compute publish` subcommand [\#242](https://github.com/fastly/cli/pull/242)
- Allow local binary to be renamed [\#240](https://github.com/fastly/cli/pull/240)
- Retain `RUSTFLAGS` values from the environment [\#239](https://github.com/fastly/cli/pull/239)
- Make GitHub Versioner configurable [\#236](https://github.com/fastly/cli/pull/236)
- Add support for `compression_codec` to logging file sink endpoints [\#190](https://github.com/fastly/cli/pull/190)

**Bug fixes:**

- Remove flaky test logic. [\#249](https://github.com/fastly/cli/pull/249)
- Check the rustup version [\#248](https://github.com/fastly/cli/pull/248)
- Print all commands and subcommands in usage [\#244](https://github.com/fastly/cli/pull/244)
- pkg/logs: fix typo in error message [\#238](https://github.com/fastly/cli/pull/238)

## [v0.26.3](https://github.com/fastly/cli/releases/tag/v0.26.3) (2021-03-26)

[Full Changelog](https://github.com/fastly/cli/compare/v0.26.2...v0.26.3)

**Enhancements:**

- Default to port 443 if UseSSL set. [\#234](https://github.com/fastly/cli/pull/234)

**Bug fixes:**

- Ensure all UPDATE operations don't set optional fields. [\#235](https://github.com/fastly/cli/pull/235)
- Avoid setting fields that cause API to fail when given zero value. [\#233](https://github.com/fastly/cli/pull/233)

## [v0.26.2](https://github.com/fastly/cli/releases/tag/v0.26.2) (2021-03-22)

[Full Changelog](https://github.com/fastly/cli/compare/v0.26.1...v0.26.2)

**Enhancements:**

- Extra error handling around loading remote configuration data. [\#229](https://github.com/fastly/cli/pull/229)

**Bug fixes:**

- `fastly compute build` exits with error 1 [\#227](https://github.com/fastly/cli/issues/227)
- Set GOVERSION for goreleaser. [\#228](https://github.com/fastly/cli/pull/228)

## [v0.26.1](https://github.com/fastly/cli/releases/tag/v0.26.1) (2021-03-19)

[Full Changelog](https://github.com/fastly/cli/compare/v0.26.0...v0.26.1)

**Bug fixes:**

- Fix manifest\_version as a section bug. [\#225](https://github.com/fastly/cli/pull/225)

## [v0.26.0](https://github.com/fastly/cli/releases/tag/v0.26.0) (2021-03-18)

[Full Changelog](https://github.com/fastly/cli/compare/v0.25.2...v0.26.0)

**Enhancements:**

- Remove version from fastly.toml manifest. [\#222](https://github.com/fastly/cli/pull/222)
- Don't run "cargo update" before building rust app. [\#221](https://github.com/fastly/cli/pull/221)

**Bug fixes:**

- Loading remote config.toml should fail gracefully. [\#223](https://github.com/fastly/cli/pull/223)
- Update the fastly.toml manifest if missing manifest\_version. [\#220](https://github.com/fastly/cli/pull/220)
- Refactor UserAgent. [\#219](https://github.com/fastly/cli/pull/219)

## [v0.25.2](https://github.com/fastly/cli/releases/tag/v0.25.2) (2021-03-16)

[Full Changelog](https://github.com/fastly/cli/compare/v0.25.1...v0.25.2)

**Bug fixes:**

- Fix duplicate warning messages and missing SetOutput\(\). [\#216](https://github.com/fastly/cli/pull/216)

## [v0.25.1](https://github.com/fastly/cli/releases/tag/v0.25.1) (2021-03-16)

[Full Changelog](https://github.com/fastly/cli/compare/v0.25.0...v0.25.1)

**Bug fixes:**

- The manifest\_version should default to 1 if missing. [\#214](https://github.com/fastly/cli/pull/214)

## [v0.25.0](https://github.com/fastly/cli/releases/tag/v0.25.0) (2021-03-16)

[Full Changelog](https://github.com/fastly/cli/compare/v0.24.2...v0.25.0)

**Enhancements:**

- Replace deprecated ioutil functions with go 1.16. [\#212](https://github.com/fastly/cli/pull/212)
- Replace TOML parser [\#211](https://github.com/fastly/cli/pull/211)
- Implement manifest\_version into the fastly.toml [\#210](https://github.com/fastly/cli/pull/210)
- Dynamic Configuration [\#187](https://github.com/fastly/cli/pull/187)

**Bug fixes:**

- Log output should be simplified when running in CI [\#175](https://github.com/fastly/cli/issues/175)
- Override error message in compute init [\#204](https://github.com/fastly/cli/pull/204)

## [v0.24.2](https://github.com/fastly/cli/releases/tag/v0.24.2) (2021-02-15)

[Full Changelog](https://github.com/fastly/cli/compare/v0.24.1...v0.24.2)

**Bug fixes:**

- Fix CI binary overlap [\#209](https://github.com/fastly/cli/pull/209)
- Fix CI workflow by switching from old syntax to new [\#208](https://github.com/fastly/cli/pull/208)
- Fix goreleaser version lookup [\#207](https://github.com/fastly/cli/pull/207)
- LogTail: Properly close response body [\#205](https://github.com/fastly/cli/pull/205)
- Add port prompt for compute init [\#203](https://github.com/fastly/cli/pull/203)
- Update GitHub Action to not use commit hash [\#200](https://github.com/fastly/cli/pull/200)

## [v0.24.1](https://github.com/fastly/cli/releases/tag/v0.24.1) (2021-02-03)

[Full Changelog](https://github.com/fastly/cli/compare/v0.24.0...v0.24.1)

**Bug fixes:**

- Logs Tail: Give the user better feedback when --from flag errors [\#201](https://github.com/fastly/cli/pull/201)

## [v0.24.0](https://github.com/fastly/cli/releases/tag/v0.24.0) (2021-02-02)

[Full Changelog](https://github.com/fastly/cli/compare/v0.23.0...v0.24.0)

**Enhancements:**

- Add static content starter kit [\#197](https://github.com/fastly/cli/pull/197)
- 🦀 Update rust toolchain [\#196](https://github.com/fastly/cli/pull/196)

**Bug fixes:**

- Fix go vet error related to missing docstring [\#198](https://github.com/fastly/cli/pull/198)

## [v0.23.0](https://github.com/fastly/cli/releases/tag/v0.23.0) (2021-01-22)

[Full Changelog](https://github.com/fastly/cli/compare/v0.22.0...v0.23.0)

**Enhancements:**

- Update Go-Fastly dependency to v3.0.0 [\#193](https://github.com/fastly/cli/pull/193)
- Support for Compute@Edge Log Tailing [\#192](https://github.com/fastly/cli/pull/192)

**Bug fixes:**

- Resolve issues with Rust integration tests. [\#194](https://github.com/fastly/cli/pull/194)
- Update URL for default Rust starter [\#191](https://github.com/fastly/cli/pull/191)

## [v0.22.0](https://github.com/fastly/cli/releases/tag/v0.22.0) (2021-01-07)

[Full Changelog](https://github.com/fastly/cli/compare/v0.21.2...v0.22.0)

**Enhancements:**

- Add support for TLS client and batch size options for splunk [\#183](https://github.com/fastly/cli/pull/183)
- Add support for Kinesis logging endpoint [\#177](https://github.com/fastly/cli/pull/177)

## [v0.21.2](https://github.com/fastly/cli/releases/tag/v0.21.2) (2021-01-06)

[Full Changelog](https://github.com/fastly/cli/compare/v0.21.1...v0.21.2)

**Bug fixes:**

- Switch from third-party dependency to our own mirror [\#184](https://github.com/fastly/cli/pull/184)

## [v0.21.1](https://github.com/fastly/cli/releases/tag/v0.21.1) (2020-12-18)

[Full Changelog](https://github.com/fastly/cli/compare/v0.21.0...v0.21.1)

**Bug fixes:**

- CLI shouldn't recommend Rust crate prerelease versions [\#168](https://github.com/fastly/cli/issues/168)
- Run cargo update before attempting to build Rust compute packages [\#179](https://github.com/fastly/cli/pull/179)

## [v0.21.0](https://github.com/fastly/cli/releases/tag/v0.21.0) (2020-12-14)

[Full Changelog](https://github.com/fastly/cli/compare/v0.20.0...v0.21.0)

**Enhancements:**

- Adds support for managing edge dictionaries [\#159](https://github.com/fastly/cli/pull/159)

## [v0.20.0](https://github.com/fastly/cli/releases/tag/v0.20.0) (2020-11-24)

[Full Changelog](https://github.com/fastly/cli/compare/v0.19.0...v0.20.0)

**Enhancements:**

- Migrate to Go-Fastly 2.0.0 [\#169](https://github.com/fastly/cli/pull/169)

**Bug fixes:**

- Build failure with Cargo workspaces [\#171](https://github.com/fastly/cli/issues/171)
- Support cargo workspaces [\#172](https://github.com/fastly/cli/pull/172)

## [v0.19.0](https://github.com/fastly/cli/releases/tag/v0.19.0) (2020-11-19)

[Full Changelog](https://github.com/fastly/cli/compare/v0.18.1...v0.19.0)

**Enhancements:**

- Support sasl kafka endpoint options in Fastly CLI [\#161](https://github.com/fastly/cli/pull/161)

## [v0.18.1](https://github.com/fastly/cli/releases/tag/v0.18.1) (2020-11-03)

[Full Changelog](https://github.com/fastly/cli/compare/v0.18.0...v0.18.1)

**Enhancements:**

- Update the default Rust template to fastly-0.5.0 [\#163](https://github.com/fastly/cli/pull/163)

**Bug fixes:**

- Constrain Version Upgrade Suggestion [\#165](https://github.com/fastly/cli/pull/165)
- Fix AssemblyScript compilation messaging [\#164](https://github.com/fastly/cli/pull/164)

## [v0.18.0](https://github.com/fastly/cli/releases/tag/v0.18.0) (2020-10-27)

[Full Changelog](https://github.com/fastly/cli/compare/v0.17.0...v0.18.0)

**Enhancements:**

- Add AssemblyScript support to compute init and build commands [\#160](https://github.com/fastly/cli/pull/160)

## [v0.17.0](https://github.com/fastly/cli/releases/tag/v0.17.0) (2020-09-24)

[Full Changelog](https://github.com/fastly/cli/compare/v0.16.1...v0.17.0)

**Enhancements:**

- Bump supported Rust toolchain version to 1.46 [\#156](https://github.com/fastly/cli/pull/156)
- Add service search command [\#152](https://github.com/fastly/cli/pull/152)

**Bug fixes:**

- Broken link in usage info [\#148](https://github.com/fastly/cli/issues/148)

## [v0.16.1](https://github.com/fastly/cli/releases/tag/v0.16.1) (2020-07-21)

[Full Changelog](https://github.com/fastly/cli/compare/v0.16.0...v0.16.1)

**Bug fixes:**

- Display the correct version number on error [\#144](https://github.com/fastly/cli/pull/144)
- Fix bug where name was not added to the manifest [\#143](https://github.com/fastly/cli/pull/143)

## [v0.16.0](https://github.com/fastly/cli/releases/tag/v0.16.0) (2020-07-09)

[Full Changelog](https://github.com/fastly/cli/compare/v0.15.0...v0.16.0)

**Enhancements:**

- Compare package hashsum during deployment [\#139](https://github.com/fastly/cli/pull/139)
- Allow compute init to be reinvoked within an existing package directory [\#138](https://github.com/fastly/cli/pull/138)

## [v0.15.0](https://github.com/fastly/cli/releases/tag/v0.15.0) (2020-06-29)

[Full Changelog](https://github.com/fastly/cli/compare/v0.14.0...v0.15.0)

**Enhancements:**

- Adds OpenStack logging support [\#132](https://github.com/fastly/cli/pull/132)

## [v0.14.0](https://github.com/fastly/cli/releases/tag/v0.14.0) (2020-06-25)

[Full Changelog](https://github.com/fastly/cli/compare/v0.13.0...v0.14.0)

**Enhancements:**

- Bump default Rust template version to v0.4.0 [\#133](https://github.com/fastly/cli/pull/133)

## [v0.13.0](https://github.com/fastly/cli/releases/tag/v0.13.0) (2020-06-15)

[Full Changelog](https://github.com/fastly/cli/compare/v0.12.0...v0.13.0)

**Enhancements:**

- Allow compute services to be initialised from an existing service ID [\#125](https://github.com/fastly/cli/pull/125)

**Bug fixes:**

- Fix bash completion [\#128](https://github.com/fastly/cli/pull/128)

**Closed issues:**

- Bash Autocomplete is broken [\#127](https://github.com/fastly/cli/issues/127)

## [v0.12.0](https://github.com/fastly/cli/releases/tag/v0.12.0) (2020-06-05)

[Full Changelog](https://github.com/fastly/cli/compare/v0.11.0...v0.12.0)

**Enhancements:**

- Adds MessageType field to SFTP [\#118](https://github.com/fastly/cli/pull/118)
- Adds User field to Cloudfiles Updates [\#117](https://github.com/fastly/cli/pull/117)
- Adds Region field to Scalyr [\#116](https://github.com/fastly/cli/pull/116)
- Adds PublicKey field to S3 [\#114](https://github.com/fastly/cli/pull/114)
- Adds MessageType field to GCS Updates [\#113](https://github.com/fastly/cli/pull/113)
- Adds ResponseCondition and Placement fields to BigQuery Creates [\#111](https://github.com/fastly/cli/pull/111)

**Bug fixes:**

- Unable to login with API key [\#94](https://github.com/fastly/cli/issues/94)

## [v0.11.0](https://github.com/fastly/cli/releases/tag/v0.11.0) (2020-05-29)

[Full Changelog](https://github.com/fastly/cli/compare/v0.10.0...v0.11.0)

**Enhancements:**

- Add ability to exclude files from build package [\#87](https://github.com/fastly/cli/pull/87)

**Bug fixes:**

- unintended files included in upload package [\#24](https://github.com/fastly/cli/issues/24)

## [v0.10.0](https://github.com/fastly/cli/releases/tag/v0.10.0) (2020-05-28)

[Full Changelog](https://github.com/fastly/cli/compare/v0.9.0...v0.10.0)

**Enhancements:**

- Adds Google Cloud Pub/Sub logging endpoint support [\#96](https://github.com/fastly/cli/pull/96)
- Adds Datadog logging endpoint support [\#92](https://github.com/fastly/cli/pull/92)
- Adds HTTPS logging endpoint support [\#91](https://github.com/fastly/cli/pull/91)
- Adds Elasticsearch logging endpoint support [\#90](https://github.com/fastly/cli/pull/90)
- Adds Azure Blob Storage logging endpoint support [\#89](https://github.com/fastly/cli/pull/89)

## [v0.9.0](https://github.com/fastly/cli/releases/tag/v0.9.0) (2020-05-21)

[Full Changelog](https://github.com/fastly/cli/compare/v0.8.0...v0.9.0)

**Breaking changes:**

- Describe subcommand consistent --name short flag -d -\> -n [\#85](https://github.com/fastly/cli/pull/85)

**Enhancements:**

- Adds Kafka logging endpoint support [\#95](https://github.com/fastly/cli/pull/95)
- Adds DigitalOcean Spaces logging endpoint support [\#80](https://github.com/fastly/cli/pull/80)
- Adds Rackspace Cloudfiles logging endpoint support [\#79](https://github.com/fastly/cli/pull/79)
- Adds Log Shuttle logging endpoint support [\#78](https://github.com/fastly/cli/pull/78)
- Adds SFTP logging endpoint support [\#77](https://github.com/fastly/cli/pull/77)
- Adds Heroku logging endpoint support [\#76](https://github.com/fastly/cli/pull/76)
- Adds Honeycomb logging endpoint support [\#75](https://github.com/fastly/cli/pull/75)
- Adds Loggly logging endpoint support [\#74](https://github.com/fastly/cli/pull/74)
- Adds Scalyr logging endpoint support [\#73](https://github.com/fastly/cli/pull/73)
- Verify fastly crate version during compute build. [\#67](https://github.com/fastly/cli/pull/67)
- Basic support for historical & realtime stats [\#66](https://github.com/fastly/cli/pull/66)
- Adds Splunk endpoint [\#64](https://github.com/fastly/cli/pull/64)
- Adds FTP logging endpoint support [\#63](https://github.com/fastly/cli/pull/63)
- Adds GCS logging endpoint support [\#62](https://github.com/fastly/cli/pull/62)
- Adds Sumo Logic logging endpoint support [\#59](https://github.com/fastly/cli/pull/59)
- Adds Papertrail logging endpoint support [\#57](https://github.com/fastly/cli/pull/57)
- Adds Logentries logging endpoint support [\#56](https://github.com/fastly/cli/pull/56)

**Bug fixes:**

- Fallback to a file copy during update if the file rename fails [\#72](https://github.com/fastly/cli/pull/72)

## [v0.8.0](https://github.com/fastly/cli/releases/tag/v0.8.0) (2020-05-13)

[Full Changelog](https://github.com/fastly/cli/compare/v0.7.1...v0.8.0)

**Enhancements:**

- Add a --force flag to compute build to skip verification steps. [\#68](https://github.com/fastly/cli/pull/68)
- Improve `compute build` rust compilation error messaging [\#60](https://github.com/fastly/cli/pull/60)
- Adds Syslog logging endpoint support [\#55](https://github.com/fastly/cli/pull/55)

**Bug fixes:**

- debian package doesn't install in default $PATH [\#58](https://github.com/fastly/cli/issues/58)
- deb and rpm packages install the binary in `/usr/local` instead of `/usr/local/bin` [\#53](https://github.com/fastly/cli/issues/53)

**Closed issues:**

- ERROR: error during compilation process: exit status 101. [\#52](https://github.com/fastly/cli/issues/52)

## [v0.7.1](https://github.com/fastly/cli/releases/tag/v0.7.1) (2020-05-04)

[Full Changelog](https://github.com/fastly/cli/compare/v0.7.0...v0.7.1)

**Bug fixes:**

- Ensure compute deploy selects the most ideal version to clone/activate  [\#50](https://github.com/fastly/cli/pull/50)

## [v0.7.0](https://github.com/fastly/cli/releases/tag/v0.7.0) (2020-04-28)

[Full Changelog](https://github.com/fastly/cli/compare/v0.6.0...v0.7.0)

**Enhancements:**

- Publish scoop package manifest during release process [\#45](https://github.com/fastly/cli/pull/45)
- Generate dep and rpm packages during release process [\#44](https://github.com/fastly/cli/pull/44)
- 🦀 🆙date to Rust 1.43.0 [\#40](https://github.com/fastly/cli/pull/40)

**Closed issues:**

- README's build instructions do not work without additional dependencies met [\#35](https://github.com/fastly/cli/issues/35)

## [v0.6.0](https://github.com/fastly/cli/releases/tag/v0.6.0) (2020-04-24)

[Full Changelog](https://github.com/fastly/cli/compare/v0.5.0...v0.6.0)

**Enhancements:**

- Bump default Rust template to v0.3.0 [\#32](https://github.com/fastly/cli/pull/32)
- Publish to homebrew [\#26](https://github.com/fastly/cli/pull/26)

**Bug fixes:**

- Don't display the fastly token in the terminal when doing `fastly configure` [\#27](https://github.com/fastly/cli/issues/27)
- Documentation typo in `fastly service-version update` [\#22](https://github.com/fastly/cli/issues/22)
- Fix typo in service-version update command [\#31](https://github.com/fastly/cli/pull/31)
- Tidy up `fastly configure` text output [\#30](https://github.com/fastly/cli/pull/30)
- compute/init: make space after Author prompt match other prompts [\#25](https://github.com/fastly/cli/pull/25)

## [v0.5.0](https://github.com/fastly/cli/releases/tag/v0.5.0) (2020-04-08)

[Full Changelog](https://github.com/fastly/cli/compare/v0.4.1...v0.5.0)

**Enhancements:**

- Add the ability to initialise a compute project from a specific branch [\#14](https://github.com/fastly/cli/pull/14)

## [v0.4.1](https://github.com/fastly/cli/releases/tag/v0.4.1) (2020-03-27)

[Full Changelog](https://github.com/fastly/cli/compare/v0.4.0...v0.4.1)

**Bug fixes:**

- Fix persistence of author string to fastly.toml [\#12](https://github.com/fastly/cli/pull/12)
- Fix up undoStack.RunIfError [\#11](https://github.com/fastly/cli/pull/11)

## [v0.4.0](https://github.com/fastly/cli/releases/tag/v0.4.0) (2020-03-20)

[Full Changelog](https://github.com/fastly/cli/compare/v0.3.0...v0.4.0)

**Enhancements:**

- Add commands for S3 logging endpoints [\#9](https://github.com/fastly/cli/pull/9)
- Add useful next step links to compute deploy [\#8](https://github.com/fastly/cli/pull/8)
- Persist version to manifest file when deploying compute services [\#7](https://github.com/fastly/cli/pull/7)

**Bug fixes:**

- Fix comment for --use-ssl flag [\#6](https://github.com/fastly/cli/pull/6)

## [v0.3.0](https://github.com/fastly/cli/releases/tag/v0.3.0) (2020-03-11)

[Full Changelog](https://github.com/fastly/cli/compare/v0.2.0...v0.3.0)

**Enhancements:**

- Interactive init [\#5](https://github.com/fastly/cli/pull/5)

## [v0.2.0](https://github.com/fastly/cli/releases/tag/v0.2.0) (2020-02-24)

[Full Changelog](https://github.com/fastly/cli/compare/v0.1.0...v0.2.0)

**Enhancements:**

- Improve toolchain installation help messaging [\#3](https://github.com/fastly/cli/pull/3)

**Bug fixes:**

- Filter unwanted files from template repository whilst initialising [\#1](https://github.com/fastly/cli/pull/1)

## [v0.1.0](https://github.com/fastly/cli/releases/tag/v0.1.0) (2020-02-05)

[Full Changelog](https://github.com/fastly/cli/compare/5a8d21b6b1973abe7a27f985856d910f4396ce95...v0.1.0)

Initial release :tada:



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*

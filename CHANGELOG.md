# Changelog

## [v0.16.1](https://github.com/fastly/cli/releases/tag/v0.16.1) (2020-07-20)

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
- Adds Kafka logging endpoint support [\#95](https://github.com/fastly/cli/pull/95)
- Adds Datadog logging endpoint support [\#92](https://github.com/fastly/cli/pull/92)
- Adds HTTPS logging endpoint support [\#91](https://github.com/fastly/cli/pull/91)
- Adds Elasticsearch logging endpoint support [\#90](https://github.com/fastly/cli/pull/90)
- Adds Azure Blob Storage logging endpoint support [\#89](https://github.com/fastly/cli/pull/89)

## [v0.9.0](https://github.com/fastly/cli/releases/tag/v0.9.0) (2020-05-21)

[Full Changelog](https://github.com/fastly/cli/compare/v0.8.0...v0.9.0)

**Breaking changes:**

- Describe subcommand consistent --name short flag -d -\> -n [\#85](https://github.com/fastly/cli/pull/85)

**Enhancements:**

- Adds DigitalOcean Spaces logging endpoint support [\#80](https://github.com/fastly/cli/pull/80)
- Adds Rackspace Cloudfiles logging endpoint support [\#79](https://github.com/fastly/cli/pull/79)
- Adds Log Shuttle logging endpoint support [\#78](https://github.com/fastly/cli/pull/78)
- Adds SFTP logging endpoint support [\#77](https://github.com/fastly/cli/pull/77)
- Adds Heroku logging endpoint support [\#76](https://github.com/fastly/cli/pull/76)
- Adds Honeycomb logging endpoint support [\#75](https://github.com/fastly/cli/pull/75)
- Adds Loggly logging endpoint support [\#74](https://github.com/fastly/cli/pull/74)
- Adds Scalyr logging endpoint support [\#73](https://github.com/fastly/cli/pull/73)
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
- Verify fastly crate version during compute build. [\#67](https://github.com/fastly/cli/pull/67)
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

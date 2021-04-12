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
Built with go version go1.16 linux/amd64
```

The Fastly CLI will notify you if a new version is available, and can update
itself via `fastly update`.

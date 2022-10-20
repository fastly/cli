#!/bin/bash
#
# DEPENDENCIES:
# cargo install tomlq

set -e

kits=(
  compute-starter-kit-rust-default
  compute-starter-kit-rust-empty
  compute-starter-kit-rust-static-content
  compute-starter-kit-rust-websockets
  compute-starter-kit-javascript-default
  compute-starter-kit-javascript-empty
  compute-starter-kit-assemblyscript-default
  compute-starter-kit-go-default
)

# The last line of the template file is still used by DevHub process
# but should be removed for the CLI's CI process.
#
# As this script will be run on both macOS and Linux (CI) we use a
# cross-platform solution for removing the last line from the template file.
# e.g. BSD sed is a different beast on macOS to the gnu version on Linux.
cat ./.fastly/config.toml | tail -r | tail -n +2 | tail -r > ./pkg/config/config.toml

function parse() {
  tomlq -f "./$k.toml" $1
}

function append() {
  echo $1 >> ./pkg/config/config.toml
}

for k in ${kits[@]}; do
  curl -s "https://raw.githubusercontent.com/fastly/$k/main/fastly.toml" -o "./$k.toml"

  append "[[starter-kits.$(parse language)]]"
  append "description = \"$(parse description)\""
  append "name = \"$(parse name)\""
  append "path = \"https://github.com/fastly/$k\""
  append ''

  rm "./$k.toml"
done

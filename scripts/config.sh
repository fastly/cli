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

# The last line of the template file is still used by an internal DevHub process
# but should be removed for the CLI's CI process.
cp ./.fastly/config.toml ./pkg/config/config.toml

if [[ $(uname) == "Darwin" ]]; then
  sed -i '' -e '$ d' ./pkg/config/config.toml
else
  sed -i '$ d' ./pkg/config/config.toml
fi

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

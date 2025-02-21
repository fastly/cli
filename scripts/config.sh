#!/usr/bin/env bash

set -e

cp ".fastly/config.toml" "pkg/config/config.toml"

if ! command -v tq &> /dev/null
then
  cargo install tomlq
fi

kits=(
  compute-starter-kit-assemblyscript-default
  compute-starter-kit-go-default
  compute-starter-kit-go-tinygo
  compute-starter-kit-javascript-default
  compute-starter-kit-javascript-empty
  compute-starter-kit-rust-default
  compute-starter-kit-rust-empty
  compute-starter-kit-rust-static-content
  compute-starter-kit-rust-websockets
  compute-starter-kit-typescript-default
)

function parse() {
  tq -r -f "$k.toml" $1
}

function append() {
  echo $1 >> pkg/config/config.toml
}

for k in ${kits[@]}; do
  curl -s "https://raw.githubusercontent.com/fastly/$k/main/fastly.toml" -o "$k.toml"

  append ''
  append "[[starter-kits.$(parse language)]]"
  append "description = \"$(parse description)\""
  append "name = \"$(parse name)\""
  append "path = \"https://github.com/fastly/$k\""

  rm "$k.toml"
done

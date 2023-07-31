#!/usr/bin/env bash
set -e

$1 help --format json > dist/usage.json

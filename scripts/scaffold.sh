#!/bin/bash
set -e

export CLI_PACKAGE=$1
export CLI_COMMAND=$2
export CLI_API=$3

mkdir -p pkg/commands/$CLI_PACKAGE

# CREATE NEW COMMAND FILES
#
cat .tmpl/test.go | envsubst > pkg/commands/$CLI_PACKAGE/${CLI_PACKAGE}_test.go
filenames=("create" "delete" "describe" "doc" "list" "root" "update")
for filename in "${filenames[@]}"; do
	cat .tmpl/$filename.go | envsubst > pkg/commands/$CLI_PACKAGE/$filename.go
done

source ./scripts/scaffold-update-interfaces.sh

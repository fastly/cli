#!/usr/bin/env bash
set -e

export CLI_CATEGORY=$1
export CLI_CATEGORY_COMMAND=$2
export CLI_PACKAGE=$3
export CLI_COMMAND=$4
export CLI_API=$5

mkdir -p pkg/commands/$CLI_CATEGORY/$CLI_PACKAGE

# CREATE NEW CATEGORY FILES
#
# NOTE: We avoid recreating the files if they already exist, which can happen
# if adding a new command to an existing category (e.g. adding a new logging
# endpoint to the logging category).
#
if [ ! -f "pkg/commands/$CLI_CATEGORY/doc.go" ]; then
	cat .tmpl/doc_parent.go | envsubst > pkg/commands/$CLI_CATEGORY/doc.go
fi
if [ ! -f "pkg/commands/$CLI_CATEGORY/root.go" ]; then
	cat .tmpl/root_parent.go | envsubst > pkg/commands/$CLI_CATEGORY/root.go
fi

# CREATE NEW COMMAND FILES
#
cat .tmpl/test.go | envsubst > pkg/commands/$CLI_CATEGORY/$CLI_PACKAGE/${CLI_PACKAGE}_test.go
filenames=("create" "delete" "describe" "doc" "list" "root" "update")
for filename in "${filenames[@]}"; do
	cat .tmpl/$filename.go | envsubst > pkg/commands/$CLI_CATEGORY/$CLI_PACKAGE/$filename.go
done

source ./scripts/scaffold-update-interfaces.sh

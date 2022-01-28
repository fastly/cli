#!/bin/bash

# The following `commits` var essentially is an array containing multiple lines
# of output, specifically one commit SHA per line.
#
# The command that generates it looks complex but in essence it's logging a
# range of commits between the latest commit and the previous tag:
#
# git log --<formatting_flags> <prev_tag>..HEAD
#
# Broken down further:
#
# git rev-list
#   = get the SHA of the commit before latest tag (the tag that triggered this release)
#
# git describe
#   = get the previous tag
#
# git log <prev_tag>..HEAD
#   = display just the SHAs and stick them in an array
#
commits=($(git log --oneline --format=format:%H $(git describe --abbrev=0 --tags $(git rev-list --tags --skip=1 --max-count=1))..HEAD))

arr="["

for commit in "${commits[@]}"; do
  arr+="{\"id\":\"$commit\"},"
done

arr+="]"

# need to remove the last trailing comma as that's invalid json and is what
# we're ultimately producing this value to be used as.
#
# so we do that using sed...

echo "::set-output name=commits::$(echo $arr | sed 's/},]/}]/')"

# example output:
#
# ::set-output name=commits::[{"id":"afe0311974601664933a37391a9f78e8e2ee320b"},{"id":"2fc2c62f9a4851a7a54af9c6fe1946f8df0a77a7"},{"id":"151c6edfff93560bcaab28acc0757a01e1eb916f"},]

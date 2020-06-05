#!/bin/bash
set -e

if ! command -v github_changelog_generator > /dev/null; then
	echo "No github_changelog_generator in \$PATH, install via 'gem install github_changelog_generator'."
	exit 1
fi

if [ -z "$CHANGELOG_GITHUB_TOKEN" ]; then
	printf "\nWARNING: No \$CHANGELOG_GITHUB_TOKEN in environment, set one to avoid hitting rate limit.\n\n"
fi

if [ -z "$SEMVER_TAG" ]; then
	echo "You must set \$SEMVER_TAG to your desired release semver version."
	exit 1
fi

github_changelog_generator -u fastly -p cli \
	--future-release $SEMVER_TAG \
	--no-pr-wo-labels \
	--no-author \
	--enhancement-label "**Enhancements:**" \
	--bugs-label "**Bug fixes:**" \
	--release-url "https://github.com/fastly/cli/releases/tag/%s" \
	--exclude-labels documentation \
	--exclude-tags-regex "v.*-.*"

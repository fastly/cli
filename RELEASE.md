# Releasing

### How to cut a new release of the CLI

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html); therefore first determine the appropriate version tag based on the change set. If in doubt discuss with the team via Slack before releasing.

1. Merge all PRs intended for the release into the `master` branch
1. Checkout and update the master branch and ensure all tests are passing:
    * `git checkout master`
    * `git pull`
    * `make all`
1. Update the [`CHANGELOG.md`](https://github.com/fastly/cli/blob/master/CHANGELOG.md):
    * Apply necessary labels (`enchancement`, `bug`, `documentation` etc) to all PRs intended for the release that you wish to appear in the `CHANGELOG.md`
    * **Only add labels for relevant changes**
    * `git checkout -b vx.x.x` where `vx.x.x` is your target version tag
    * `make changelog`
    * `git add CHANGELOG.md && git commit -m "vX.X.X"`
1. Send PR for the `CHANGELOG.md`
1. Once approved and merged, checkout and update the `master` branch:
    * `git checkout master`
    * `git pull`
1. Create a new tag for `master`:
    * `git tag -s vx.x.x -m "vx.x.x"`
1. Push the new tag:
    * `git push origin vx.x.x`
1. Go to GitHub and check that the release was successful:
    * Check the release CI job status via the [Actions](https://github.com/fastly/cli/actions?query=workflow%3ARelease) tab
    * Check the release exists with valid assets and changelog: https://github.com/fastly/cli/releases
1. Announce release internally via Slack
1. Celebrate :tada:

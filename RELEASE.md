# Release Process

1. Merge all PRs intended for the release.
1. Ensure any relevant `FIXME` notes in the code are addressed (e.g. `FIXME: remove this feature before next major release`).
1. Rebase latest remote main branch locally (`git pull --rebase origin main`).
1. Ensure all analysis checks and tests are passing (`time TEST_COMPUTE_INIT=1 TEST_COMPUTE_BUILD=1 TEST_COMPUTE_DEPLOY=1 make all`).
1. Ensure goreleaser builds locally (`make release GORELEASER_ARGS="--snapshot --skip=validate --skip=post-hooks --clean"`).
1. Open a new PR to update CHANGELOG ([example](https://github.com/fastly/cli/pull/273))<sup>[1](#note1)</sup>.
1. Merge CHANGELOG.
1. Rebase latest remote main branch locally (`git pull --rebase origin main`).
1. Tag a new release (`tag=vX.Y.Z && git tag -s $tag -m "$tag" && git push origin $tag`)<sup>[2](#note2)</sup>.
1. Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/cli/releases).
1. Publish draft release.

## Creation of npm modules

Each release of the Fastly CLI will trigger the creation of a new version of the `@fastly/cli` npm module, as well as multiple per-platform modules (e.g. `@fastly/cli-darwin-arm64`). The root module (`@fastly/cli`) depends on the platform-specific modules as `optionalDependencies` so that npm will only install the compatible platform-specific package.

To see an example of the module layout, run...

$ npm install @fastly/cli-darwin-arm64 --verbose

Next, run...

$ ls node_modules/@fastly/cli-darwin-arm64

You should see a `fastly` executable binary as well as an `index.js` shim which allows the package to be imported as a module by other JavaScript projects.

## Footnotes

1. <a name="note1"></a>We utilize [semantic versioning](https://semver.org/) and only include relevant/significant changes within the CHANGELOG (be sure to document changes to the app config if `config_version` has changed, and if any breaking interface changes are made to the fastly.toml manifest those should be documented on https://fastly.com/documentation/developers).
1. <a name="note2"></a>Triggers a [github action](https://github.com/fastly/cli/blob/main/.github/workflows/tag_release.yml) that produces a 'draft' release.

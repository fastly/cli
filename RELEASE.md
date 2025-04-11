# Release Process

1. Merge all PRs intended for the release.
1. Ensure any relevant `FIXME` notes in the code are addressed (e.g. `FIXME: remove this feature before next major release`).
1. Rebase latest remote main branch locally (`git pull --rebase origin main`).
1. Ensure all analysis checks and tests are passing (`time TEST_COMPUTE_INIT=1 TEST_COMPUTE_BUILD=1 TEST_COMPUTE_DEPLOY=1 make all`).
1. Ensure goreleaser builds locally (`make release GORELEASER_ARGS="--snapshot --skip=validate --skip=post-hooks --clean"`).
1. Open a new PR to update CHANGELOG ([example](https://github.com/fastly/cli/pull/273)).
    - We utilize [semantic versioning](https://semver.org/) and only include relevant/significant changes within the CHANGELOG (be sure to document changes to the app config if `config_version` has changed, and if any breaking interface changes are made to the fastly.toml manifest those should be documented on https://fastly.com/documentation/developers).
1. Merge CHANGELOG.
1. Rebase latest remote main branch locally (`git pull --rebase origin main`).
1. Create a new signed tag (replace `{{remote}}` with the remote pointing to the official repository i.e. `origin` or `upstream` depending on your Git workflow): `tag=vX.Y.Z && git tag -s $tag -m $tag && git push {{remote}} $tag`
    - Triggers a [github action](https://github.com/fastly/cli/blob/main/.github/workflows/tag_release.yml) that produces a 'draft' release.
1. Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/cli/releases).
1. Publish draft release.

## Creation of npm packages

Each release of the Fastly CLI triggers a workflow in `.github/workflows/publish_release.yml` that results in the creation of a new version of the `@fastly/cli` npm package, as well as multiple packages each representing a supported platform/arch combo (e.g. `@fastly/cli-darwin-arm64`). These packages are given the same version number as the Fastly CLI release. The workflow then publishes the `@fastly/cli` package and the per-platform/arch packages to npmjs.com using the `NPM_TOKEN` secret in this repository. The per-platform/arch packages are generated on each release and not committed to source control.

> [!NOTE]
> The workflow step that performs `npm version` in the directory of the `@fastly/cli` package triggers the execution of the `version` script listed in its `package.json`. In turn, this script creates the per-platform/arch packages.

The `@fastly/cli` package is set up to declare the platform/arch-specific packages as `optionalDependencies`. When a package installs `@fastly/cli` as one of its `dependencies`, npm will additionally install just the platform/arch-specific package compatible with the environment.

> [!NOTE]
> The `optionalDependencies` list only restricts the packages that are actually installed into the `node_modules` directory in an environment, and does not affect what is saved to the lockfile (`package-lock.json`). All the platform/arch-specific packages will be listed in the lockfile, so a single lockfile is safe to use in environments that may represent a different platform/arch combo.

To see an example of the module layout, run:

```sh
npm install @fastly/cli-darwin-arm64 --verbose
ls node_modules/@fastly/cli-darwin-arm64
```

You should see a `fastly` executable binary as well as an `index.js` shim which allows the package to be imported as a module by other JavaScript projects.

# Release Process

> ⚠️ **IMPORTANT:** If publishing a new major version, ensure the module name in
> the [go.mod](./go.mod) (and any references to the module name in the code) are
> updated to reflect the latest release.

## Example

```diff
- module github.com/fastly/cli/v1
+ module github.com/fastly/cli/v2

...

- import "github.com/fastly/cli/v1/pkg/app"
+ import "github.com/fastly/cli/v2/pkg/app"
```

## Steps

- Ensure any relevant `FIXME` notes are resolved.
- Merge all PRs intended for the release.
- Rebase latest remote changes:

```shell
git pull --rebase origin main
```

- Ensure all analysis checks and tests are passing:

```shell
time TEST_COMPUTE_INIT=1 TEST_COMPUTE_BUILD=1 TEST_COMPUTE_DEPLOY=1 make all
```

- Ensure goreleaser builds locally:

```shell
make release GORELEASER_ARGS="--skip=validate --skip=post-hooks --clean"
```

- Open a new PR to update [CHANGELOG](./CHANGELOG.md).

> **NOTE:** Document changes to the app config (if `config_version` has
> changed), and any changes to the [fastly.toml](https://www.fastly.com/documentation/reference/compute/fastly-toml/).

- Merge CHANGELOG.
- Rebase latest remote changes.
- Tag a new release:

```shell
tag=vX.Y.Z && git tag -s $tag -m "$tag" && git push origin $tag
```

> **NOTE:** This will trigger a [github action](https://github.com/fastly/cli/blob/main/.github/workflows/tag_release.yml)
> that produces a 'draft' release.

- Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/cli/releases).
- Publish draft release.

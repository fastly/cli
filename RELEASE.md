# Release Process

1. Merge all PRs intended for the release.
2. Rebase latest remote main branch locally (`git pull --rebase origin main`).
3. Ensure all analysis checks and tests are passing (`time TEST_COMPUTE_INIT=1 TEST_COMPUTE_BUILD=1 TEST_COMPUTE_DEPLOY=1 make all`).
4. Ensure goreleaser builds locally (`make fastly GORELEASER_ARGS="--skip-validate --clean"`).
5. Open a new PR to update CHANGELOG ([example](https://github.com/fastly/cli/pull/273))<sup>[1](#note1)</sup>.
6. Merge CHANGELOG.
7. Rebase latest remote main branch locally (`git pull --rebase origin main`).
8. Tag a new release (`tag=vX.Y.Z && git tag -s $tag -m "$tag" && git push origin $tag`)<sup>[2](#note2)</sup>.
9. Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/cli/releases).
10. Publish draft release.
11. Communicate the release in the relevant Slack channels<sup>[3](#note3)</sup>.

## Footnotes

1. <a name="note1"></a>We utilize [semantic versioning](https://semver.org/) and only include relevant/significant changes within the CHANGELOG.
2. <a name="note2"></a>Triggers a [github action](https://github.com/fastly/cli/blob/main/.github/workflows/tag_release.yml) that produces a 'draft' release.
3. <a name="note3"></a>Fastly make internal announcements in the Slack channels: `#api-clients`, `#ecp-languages`.

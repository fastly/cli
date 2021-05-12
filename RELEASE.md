# Release Process

1. Merge all PRs intended for the release.
2. Rebase latest remote main branch locally (`git pull --rebase origin main`).
3. Ensure all analysis checks and tests are passing (`make all`).
4. Open a new PR to update CHANGELOG ([example](https://github.com/fastly/cli/pull/273)).
5. Merge CHANGELOG.
6. Rebase latest remote main branch locally (`git pull --rebase origin main`).
7. Tag a new release (`git tag -s vX.Y.Z -m "vX.Y.Z" && git push origin vX.Y.Z`).
8. Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/cli/releases).
9. Publish draft release.
10. Communicate the release in the relevant Slack channels.

## Notes

Step 4. we utilize [semantic versioning](https://semver.org/) and only include relevant/significant changes within the CHANGELOG.

Step 7. triggers a [github action](https://github.com/fastly/cli/blob/main/.github/workflows/tag_release.yml) that produces a 'draft' release.

Step 10. `#api-clients`, `#ecp-languages`.

# Contributing

We're happy to receive feature requests and PRs that address real problems
users run into while using the CLI. If your change is nontrivial, please open an
[issue](https://github.com/fastly/cli/issues/new) to discuss the idea and
implementation strategy before submitting a PR.

1. Fork the repository.
2. Create an `upstream` remote.
    ```bash
    $ git remote add upstream git@github.com:fastly/cli.git
    ```
3. Create a feature branch.
4. Write tests.
5. Run the linter and formatter `make all`. 
    1. You may need to install [golangci-lint](https://golangci-lint.run/welcome/install/) if you don't have it installed
6. Add your changes to `CHANGELOG.md` in [Commitizen](https://commitizen-tools.github.io/commitizen/) style message
7. Open a pull request against `upstream main`.
    1. Fill out the [pull request template](./.github/PULL_REQUEST_TEMPLATE.md). PRs that are ready for review but do not follow the template are labeled `needs-template` and closed automatically after 3 days.
    2. Once you have marked your PR as `Ready for Review` please do not force push to the branch
8. Celebrate :tada:!

# Contributing

We're happy to receive feature requests and PRs. If your change is nontrivial,
please open an [issue](https://github.com/fastly/cli/issues/new) to discuss the
idea and implementation strategy before submitting a PR.

1. Fork the repository.
1. Create an `upstream` remote.
```bash
$ git remote add upstream git@github.com:fastly/cli.git
```
1. Create a feature branch.
1. Write tests.
1. Validate and prepare your change.
```bash
$ make all
```
1. Open a pull request against `upstream master`.
1. Celebrate :tada:!

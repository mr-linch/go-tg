# Contributing to go-tg

Thanks for taking the time to improve go-tg!

This document describes how to prepare a PR for a change in the main repository.

- [Prerequisites](#prerequisites)
- [Making changes](#making-changes)
- [Submit changes](#submit-changes)

## Prerequisites

- Go 1.18+

If you haven't already, you can fork the main repository and clone your fork so that you can work locally:

```
git clone https://github.com/your_username/go-tg.git
```

> It is recommended to create a new branch from master for each of your bugfixes and features.
> This is required if you are planning to submit multiple PRs in order to keep the changes separate for review until they eventually get merged.

## Making changes

**Before making a PR to the main repository, it is a good idea to:**

- Add unit tests for your changes (we are using the standard [`testing`](https://pkg.go.dev/testing) go package and [`testify`](https://github.com/stretchr/testify) as assert lib).
  To run the tests, you could execute (while in the root project directory):

  ```sh
  go test -v -race ./...
  ```

- Run the linter - **golangci** ([see how to install](https://golangci-lint.run/usage/install/#local-installation)):

  ```sh
  golangci-lint run  ./...
  ```

## Submit Changes

After the change is ready, create a draft pull request, wait until all checks have passed, and also check the Codecov report for degradation of coverage by tests. After all the checks are passed, move the pull request from draft to ready and request review.

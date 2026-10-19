# Contributing to webtest

Thanks for your interest in improving `webtest`. Bug reports, feature ideas, documentation fixes, and pull requests are all welcome.

## Before you start

- For anything larger than a small fix, please [open an issue](https://github.com/lechefran/webtest/issues/new/choose) first so we can agree on the approach.
- Security issues must **not** be reported in public issues. See [SECURITY.md](SECURITY.md).
- By participating you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Development setup

You need Go 1.27 or later. The exact version is in [go.mod](go.mod).

```sh
git clone https://github.com/lechefran/webtest.git
cd webtest
go test ./...
```

## Checks

CI runs the following checks on every pull request. Please run them locally before you push.

```sh
gofmt -l .                     # must print nothing
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
go test -race ./...            # -race needs cgo; it is optional on Windows
```

## Pull requests

1. Branch from `main`.
2. Keep each change focused. Add or update tests for any behavior change.
3. Document any new or changed exported API with a godoc comment, and add an `Example` in [example_test.go](example_test.go) when it helps.
4. Add an entry under `## [Unreleased]` in [CHANGELOG.md](CHANGELOG.md). Mark breaking changes with **Breaking:**.
5. Write commit messages in the imperative mood, for example `add retry option` or `fix nil response handling`.
6. Open the PR and fill in the template. CI must pass before merge.

## Releases

Maintainers release by tagging `main` with a semantic version, for example `git tag -a v0.6.0 -m "v0.6.0"` followed by `git push origin v0.6.0`. The release workflow publishes a GitHub Release and makes the version available on the Go module proxy and pkg.go.dev.

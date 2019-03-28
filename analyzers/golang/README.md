# Go license analyzer

Will run for the local repository source given and searches for all (recursively) imported dependencies licenses.

## Detect

Detect function decides if the given repository uses Go dependencies or not.

This is done by searching for the used dependency manager's dependency list descriptor files.
Currently we use [Dep](https://github.com/golang/dep), previously we used [Godep](https://github.com/tools/godep) for Go dependency management, the descriptor files are `Gopkg.toml`, `Godeps.json` and `go.mod`.

The analyzer is not searching dependency list descriptor files inside the vendor directory.

## Analyze

Analyze functions collects the licenses used by the given project's dependencies.

The function uses a Go package: https://github.com/godrei/licenses, which is a fork of https://github.com/pmezard/licenses, which repository is ported from https://github.com/benbalter/licensee.

The `github.com/godrei/licenses` package uses `go list` command to list the packages in a go project directory, then uses another `go list` command call to determine all (recursively) imported dependencies for each package.

Then for every imported dependency, the package looks for license files in the dependency package import path, and down to
parent directories until a file is found or `$GOPATH/src` is reached. The license files are [scored](https://github.com/godrei/licenses/blob/master/licenses/licenses.go#L335) based on how likely they are matching to license file names.

Then the license files are matched to [license templates](https://github.com/godrei/licenses/tree/master/assets) and are [scored](https://github.com/godrei/licenses/blob/master/licenses/licenses.go#L133).
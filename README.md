# GitHub License Collector

Install:

```
go install github.com/bitrise-io/github-license-collector
```

Usage:

You have to set a GitHub Personal Access Token (https://github.com/settings/tokens), so that the tool can access private repos of the specified org:

```
export GITHUB_PERSONAL_ACCESS_TOKEN='..Personal Access Token..'
```

Now you can run the `collect` command to collect all the repos of a specified GitHub Organization and then collect the licenses for all of those repos:

```
github-license-collector collect --org=GitHubOrgName
```

### Requirements
- Node 10, `brew install node@10`
- Yarn, `npm install -g yarn`
- Get go packages, `go get ./...`

### Analyzers

Each analyzer has a **Detect** logic in which it can decide if the given repository source is matching for the type it is looking for. The detected phase's description can be found in each analyzer's readme.

Also all analyzer have an **AnalyzeRepository** logic which will try to fetch all the dependencies from the project and searches for its license type. They way it reads the list of dependencies is documented in the analyzers readme.

- [NPM License Analyzer](/analyzers/npm)
- [Go License Analyzer](/analyzers/golang)
- [Ruby License Analyzer](/analyzers/ruby)
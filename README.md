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
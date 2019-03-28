# NPM license analyzer

Will run for the local repository source given.

## Detect

Checks if there is a `package.json` or `yarn.lock` in the repository, will mean detected if found any.

## Analyze

Will call `yarn licenses list --json --no-progress` command which returns the list of dependencies with its license info.

The official definition for this command can be found here: https://yarnpkg.com/lang/en/docs/cli/licenses.
Calling this command will install the dependencies used in the project automatically.
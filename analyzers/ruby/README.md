# Ruby (Bundler) license analyzer

Will run for the local repository source given.
Uses ruby and bundler as a dependency.

## Detect

Checks if there is a `Gemfile` in the repository, will mean detected if found any.

## Analyze

Collects all `Gemfile.lock` files in the repository (outside of a `vendor` directory).
Uses the bundler lockfile parser from a ruby script to get a list of dependencies.
It assumes all dependencies are from rubygems.org. Queries the rubygem API (https://rubygems.org/api/v1/versions/[depname].json" and gets the `licenses` key of the first element (latest version) of the returned array.

One package can be made available under multiple licenses. Meaning we, the user can choose the less restrictive of the licenses e.g. MIT instead of GPL.
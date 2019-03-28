# Ruby (Bundler) license analyzer

Will run for the local repository source given.

## Detect

Checks if there is a `Gemfile` in the repository, will mean detected if found any.

## Analyze

Collects all `Gemfile.lock` files in the repository (outside of a `vendor` directory).
Uses the bundler lockfile parser from a ruby script to get a list of dependencies.
It assumes all dependencies are from rubygems.org. Queries the rubygem API (https://rubygems.org/api/v1/versions/[depname].json" and gets the `licenses` key of the first element (latest version) of the returned array.
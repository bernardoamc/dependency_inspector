# Dependency Confusion

A small Go program aiming to provide insights into `Gemfile.lock` or `yarn.lock` files and their dependencies.

### Commands

#### analyze

This option is useful to detect inconsistent dependencies and hopefully prevent supply chain attacks.

The follow algorithm is used:

1. Parse every `.lock` file found in the specified directory
2. Parse the `registry` file containing the dependencies and the URL they should be loaded from
3. For every `.lock` file, flag dependencies that exists in the `registry`` but that fetched from a different URL

See the `Registry file format` section for more information.

#### remotes

This option is useful to detect suspicious remotes being used across your `.lock` files.

The follow algorithm is used:

1. Parse every `.lock` file found in the specified directory
2. Aggregate every remote found in the dependencies
3. Print the unique remotes found

For example, if dependencies are all from `https://registry.yarnpkg.com` it will print a single line with that information, ignoring dependency names.

### Registry file format

```json
{
	"url": "https://npm.acme.io/",
	"dependencies": [
        "dependency1",
        "dependency2",
        "dependencyN"
    ]
}
```

### Examples

Keep the following things in mind besides the examples shared below:

1. the `--path` flag can also point to a single `.lock` file instead of a directory.
2. the `--verbose` flag will print the dependencies parsed from each `.lock` file.

Use the `--help` flag to get more information about the available options.

#### Analyzing Gemfile.lock files for incorrect dependencies

Parsing every `.lock` file found in `lock_files/ruby` and checking if any of the dependencies are being loaded from a different `url` than the one specified in `registries/ruby.json`. For example, if we had the following `.lock` files:

- `lock_files/ruby/repo1.lock`
- `lock_files/ruby/repo2.lock`
- `lock_files/ruby/repo3.lock`

And our program found inconsistencies in `repo1` and `repo3`, the following files would be created:

- `analyze_output/repo1.json`
- `analyze_output/repo3.json`

```bash
./dependency_inspector analyze --path lock_files/ruby --registry registries/ruby.json --ruby
```

#### Analyzing yarn.lock files for incorrect dependencies

```bash
./dependency_inspector analyze --path lock_files/js --registry registries/js.json --js
```

#### Listing unique remotes from every Gemfile.lock file in directory

```bash
./dependency_inspector remotes --path lock_files/ruby --ruby
```

We can also `grep` by a particular substring:

```bash
./dependency_inspector remotes --path lock_files/ruby --grep "acme" --ruby
```

#### Listing unique remotes from every yarn.lock file in directory

```bash
./dependency_inspector remotes --path lock_files/js --js
```

The `grep` flag is also available for the `js` option.

### Scripts

The `scripts` directory contains a few scripts that can be used to fetch `.lock` files from your repositories and perform a partial analysis from dependencies within your `registry` file against `rubygems`.

### Known limitations (for now)

`yarn.lock` files with entries that contain dependencies with distinct names in the same line tend to cause duplicated remote entries in the output.

Example:

```lock
"wrap-ansi-cjs@npm:wrap-ansi@^7.0.0", wrap-ansi@^7.0.0:
```

This will return the following remote:

- https://registry.npmjs.org/wrap-ansi/-/wrap-ansi-7.0.0.tgz

Instead of the expected:

- https://registry.npmjs.org/

This is happening because our parser will fetch `wrap-ansi-cjs` from the definition above and won't be able to clean the remote URL since it contains `wrap-ansi`.

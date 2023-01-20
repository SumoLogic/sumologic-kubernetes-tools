# update-collection-v3

`update-collection-v3` helps users migrate their `values.yaml` files from v2 to v3
of [`sumologic-kubernetes-collection`][collection_github].

[collection_github]: https://github.com/SumoLogic/sumologic-kubernetes-collection

## How to build

```bash
make build
```

## How to run

```bash
./update-collection-v3 -in values.yaml -out my_new_values.yaml
```

## How to add a new migration

1. Create a new directory in the `migrations` directory, e.g.:

   ```bash
   mkdir migrations/encapsulate-otellogs-config
   ```

1. Create a `migrate.go` file in the directory and in it create an exported function `func Migrate(string) (string, error)`.

1. Add the name of the directory and the created migration function to the `migrationDirectoriesAndFunctions` variable
   in [main.go](main.go) file.

1. Create input and output test files in `testdata` subdirectory of your migration directory.
   These should act as unit tests for your migration.

1. Add or update the integration test files in the root [testdata](testdata/) directory to take your migration into account.
   These tests act as integration tests for all migrations together.

1. Implement the migration in the function as desired.

## `values.yaml` schema

[`valuesV2`][valuesV2] structure was generated using [yaml-to-go][yaml-to-go].

[yaml-to-go]: https://zhwt.github.io/yaml-to-go/
[valuesv2]: ./valuesv2.go

## Known issues

1. This package is using github.com/go-yaml/yaml which unfortunately doesn't allow
   to maintain `yaml` comments and order when using user defined structures.

   This could be done when we'd use [`yaml.Node`][yaml_node] instead of customized structs
   which reflect the schema of `values.yaml` used in `sumologic-kubernetes-collection`
   but then struct manipulation would be much more complicated.

   There is a PR pending that would add that functionality but unfortunately `go-yaml`
   doesn't seem to be actively maintained.

[yaml_node]: https://pkg.go.dev/gopkg.in/yaml.v3#Node
[pr_726]: https://github.com/go-yaml/yaml/pull/726

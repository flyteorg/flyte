# Dolt

```{eval-rst}
.. tags:: Integration, Data, SQL, Intermediate
```

```{image} https://img.shields.io/badge/Blog-Dolt-blue?style=for-the-badge
:target: https://blog.flyte.org/upleveling-flyte-data-lineage-using-dolt
:alt: Dolt Blog Post
```

The `DoltTable` plugin is a wrapper that uses [Dolt](https://github.com/dolthub/dolt) to move data between
`pandas.DataFrame`'s at execution time and database tables at rest.

## Installation

The dolt plugin and dolt command line tool are required to run these examples:

```bash
pip install flytekitplugins.dolt
sudo bash -c 'curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | sudo bash'
```

Dolt requires a user configuration to run `init`:

```bash
dolt config --global --add user.email <email>
dolt config --global --add user.name <name>
```

These demos assume a `foo` database has been created locally:

```bash
mkdir foo
cd foo
dolt init
```

```{auto-examples-toc}
dolt_quickstart_example
dolt_branch_example
```

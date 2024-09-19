# Flytekit Dolt Plugin

The DoltTable plugin is a wrapper that uses [Dolt](https://github.com/dolthub/dolt) to move data between pandas.DataFrame’s at execution time and database tables at rest.

The Dolt plugin and Dolt command-line tool can be installed as follows:

```bash
pip install flytekitplugins.dolt
sudo bash -c 'curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | sudo bash'
```

Dolt requires a user configuration to run init:

```
dolt config --global --add user.email <email>
dolt config --global --add user.name <name>
```

All the [examples](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/flytekit_plugins/dolt/index.html) can be found in the documentation.

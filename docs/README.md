# Building the Flyte docs locally

## Prerequisites

* `conda` (We recommend Miniconda installed with an [official installer](https://docs.conda.io/projects/miniconda/en/latest/index.html#latest-miniconda-installer-links))

* [`conda-lock`](https://github.com/conda/conda-lock)


## Set up the build environment

In the `flyteorg/flyte` root directory do:

```bash
$ conda-lock install --name monodocs-env monodocs-environment.lock.yaml
$ conda activate monodocs-env
$ pip install ./flyteidl
```

This creates a new environment called `monodocs-env` with all the dependencies needed to build the docs. You can choose a different environment name if you like.


## Building the docs

In the `flyteorg/flyte` root directory make sure you have activated the `monodocs-env` (or whatever you called it) environment and do:

```bash
$ make docs
```

The resulting `html` files will be in `docs/_build/html`.

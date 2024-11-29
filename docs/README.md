# Building the Flyte docs locally

## Prerequisites

* `uv` 

For MacOS / Linux:

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

For Windows, use irm to download the script and execute it with iex:

```bash
powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 | iex"
powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/0.5.5/install.ps1 | iex"
```

To install via PyPi:

```bash 
pipx install uv
```

or  

```bash 
 pip install uv
```


## Setup the build environment

```bash
uv venv
source .venv/bin/activate
uv sync
uv pip install ./flyteidl
```

## Building the docs

```bash
export DOCSEARCH_API_KEY=fake-api-key
export FLYTEKIT_LOCAL_PATH=<PATH_TO_LOCAL_FLYTEKIT_REPO>
export FLYTESNACKS_LOCAL_PATH=<PATH_TO_LOCAL_FLYTESNACKS_REPO>
uv run make docs
```
The resulting `html` files will be in `docs/_build/html`.
# flytectl

[![Docs](https://readthedocs.org/projects/flytectl/badge/?version=latest&style=plastic)](https://flytectl.rtfd.io)
[![Current Release](https://img.shields.io/github/release/flyteorg/flytectl.svg)](https://github.com/flyteorg/flytectl/releases/latest)
![Master](https://github.com/flyteorg/flytectl/workflows/Master/badge.svg)
[![GoDoc](https://godoc.org/github.com/flyteorg/flytectl?status.svg)](https://pkg.go.dev/mod/github.com/flyteorg/flytectl)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![CodeCoverage](https://img.shields.io/codecov/c/github/flyteorg/flytectl.svg)](https://codecov.io/gh/flyteorg/flytectl)
[![Go Report Card](https://goreportcard.com/badge/github.com/flyteorg/flytectl)](https://goreportcard.com/report/github.com/flyteorg/flytectl)
![Commit activity](https://img.shields.io/github/commit-activity/w/lyft/flytectl.svg?style=plastic)
![Commit since last release](https://img.shields.io/github/commits-since/lyft/flytectl/latest.svg?style=plastic)

Flytectl is designed to be a portable, lightweight, CLI for working with Flyte.  It is written in Golang and can access FlyteAdmin

## Docs

Docs are generated using Sphinx and are available at [flytectl.rtfd.io](https://flytectl.rtfd.io)
Generating docs locally can be accomplished by running make gendocs from within the docs folder


## Installation

```bash
$ brew tap flyteorg/homebrew-tap
$ brew install flytectl
```
## Get Started 

### Create a sandbox cluster 
```bash
$ docker run --rm --privileged -p 30081:30081 -p 30082:30082 -p 30084:30084 ghcr.io/flyteorg/flyte-sandbox
```

### Register your first workflow

```bash
# Run Core workflows 
$ flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.89/flytesnacks-core.tgz -d development  -p flytesnacks --archive  
# You can find all example at flytesnacks release page  https://github.com/flyteorg/flytesnacks/releases/tag/v0.2.89 
```
## Contributing

[Contribution guidelines for this project](docs/CONTRIBUTING.md)

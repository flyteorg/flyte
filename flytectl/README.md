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

Install flytectl with homebrew tap
```bash
$ brew install flyteorg/homebrew-tap/flytectl

# Upgrade flytectl 
$ brew upgrade flytectl
```

Install flytectl with shell script
```bash
$ curl -s https://raw.githubusercontent.com/lyft/flytectl/master/install.sh | bash
```

## Get Started 

### Create a sandbox cluster 
```bash
$ flytectl sandbox start 
```

### Register flytesnacks example
```bash
# Run Core workflows 
$ flytectl register examples -d development -p flytesnacks
```

## Contributing

[Contribution guidelines for this project](docs/CONTRIBUTING.md)

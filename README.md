Common Go Tools
=====================
[![Current Release](https://img.shields.io/github/release/lyft/flytestdlib.svg)](https://github.com/lyft/flytestdlib/releases/latest)
[![Build Status](https://travis-ci.org/lyft/flytestdlib.svg?branch=master)](https://travis-ci.org/lyft/flytestdlib)
[![GoDoc](https://godoc.org/github.com/lyft/flytestdlib?status.svg)](https://godoc.org/github.com/lyft/flytestdlib)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)

Shared components we found ourselves building time and time again, so we collected them in one place!

This library consists of:
 - config

   Enables strongly typed config throughout your application. Offers a way to represent config in go structs. takes care of parsing, validating and watching for changes on config.

 - cli/pflags

   Tool to generate a pflags for all fields in a given struct.

 - storage

   Abstract storage library that uses stow behind the scenes to connect to s3/azure/gcs but also offers configurable factory, in-memory storage (for testing) as well as native protobuf support.

 - contextutils

   Wrapper around golang's context to set/get known keys.

 - logger

   Wrapper around logrus that's configurable, taggable and context-aware.

 - profutils

   Starts an http server that serves /metrics (exposes prometheus metrics), /healthcheck and /version endpoints.

 - promutils

   Exposes a Scope instance that's a more convenient way to construct prometheus metrics and scope them per component.

 - atomic

   Wrapper around sync.atomic library to offer AtomicInt32 and other convenient types.

 - sets

   Offers strongly types and convenient interface sets.

 - utils
 - version

Contributing
------------

## Versioning

This repo follows [semantic versioning](https://semver.org/).

## Releases

This repository is hooked up with [goreleaser](https://goreleaser.com/). Maintainers are expected to create tags and let goreleaser compose the release message and create a release.

To create a new release, follow these steps:

- Create a PR with your changes.

- [Optional] Create an alpha tag on your branch and push that.
  
  - First get existing tags `git describe --abbrev=0 --tags`
  
  - Figure out the next alpha version (e.g. if tag is v1.2.3 then you should create a v1.2.4-alpha.0 tag)
  
  - Create a tag `git tag v1.2.4-alpha.0`
  
  - Push tag `git push --tags`

- Merge your changes and checkout master branch `git checkout master && git pull`

- Bump version tag and push to branch.

  - First get existing tags `git describe --abbrev=0 --tags`
  
  - Figure out the next release version (e.g. if tag is v1.2.3 then you should create a v1.2.4 tag or v1.3.0 or a v2.0.0 depending on what has changed. Refer to [Semantic Versioning](https://semver.org/) for information about when to bump each)
  
  - Create a tag `git tag v1.2.4`
  
  - Push tag `git push --tags`


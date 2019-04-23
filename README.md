Common Go Tools
=====================
[![Current Release](https://img.shields.io/github/release/flytestdlib.svg)](https://github.com/lyft/flytestdlib/releases/latest)
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

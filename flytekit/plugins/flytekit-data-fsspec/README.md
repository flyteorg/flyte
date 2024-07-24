# fsspec data plugin for Flytekit — Experimental

This plugin provides an implementation of the data persistence layer in Flytekit that uses fsspec. Once this plugin
is installed, it overrides all default implementations of the data plugins and provides the ones supported by fsspec. This plugin
will only install the fsspec core. To install all fsspec plugins, please follow the [fsspec documentation](https://filesystem-spec.readthedocs.io/en/latest/).

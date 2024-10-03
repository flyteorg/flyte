## Contributing to FlyteIDL

## Install Tooling

1. Run ``make download_tooling`` to install generator dependencies.

```bash
   make download_tooling
```

2. Ensure Docker is installed locally.
3. Run ``make generate`` to generate all the code, mock client, and docs for FlyteAdmin Service.

```bash
    make generate
```

4. To add new dependencies for documentation generation, modify ``doc-requirements.in`` and run

```bash
   make doc-requirements.txt
```

## Docs structure

The index.rst files for protos are arranged in parallel under the ``docs`` folder.
All the proto definitions are within ``protos/flyteidl`` and their corresponding docs are in ``protos/docs``.

```
docs
├── admin
│   ├── admin.rst
│   └── index.rst
├── core
│   ├── core.rst
│   └── index.rst
├── datacatalog
│   ├── datacatalog.rst
│   └── index.rst
├── event
│   ├── event.rst
│   └── index.rst
├── plugins
│   ├── index.rst
│   └── plugins.rst
├── service
│   ├── index.rst
│   └── service.rst
```

Each module in protos has a module in docs with the same name.
For example: ``protos/flyteidl/core`` has a module ``protos/docs/core`` under the ``docs`` folder which has the corresponding index and documentation files.


## Generating Documentation

* If a new module is to be introduced, follow the structure for core files in `generate_protos.sh` file which helps generate the core documentation from its proto files.
```
     core_proto_files=`ls protos/flyteidl/core/*.proto |xargs`
     # Remove any currently generated file
     ls -d protos/docs/core/* | grep -v index.rst | xargs rm
     protoc --doc_out=protos/docs/core --doc_opt=restructuredtext,core.rst -I=protos `echo $core_proto_files`
```

* ``make generate`` generates the modified rst files.

* ``make html`` generates the Sphinx documentation from the docs folder that uses the modified rst files.


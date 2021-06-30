# Flyte IDL

This is one of the core repositories of Flyte and contains the Specification of the Flyte Lanugage using protobuf messages, the Backend API specification in gRPC and Swagger REST. The repo contains generated clients and protocol message structures in multiple languages. Along-with the generated code, the repository also contains the Golang clients for Flyte's backend API's (the services grouped under Flyteadmin).

* [flyte.org](https://flyte.org)
* [Flyte Docs](http://docs.flyte.org)
* [FlyteIDL API reference documentation](https://docs.flyte.org/projects/flyteidl/en/stable/index.html)

## Contributing to FlyteIDL

## Tooling for FlyteIDL

1. Run ``make download_tooling`` to install generator dependencies

```bash
   make download_tooling
```

2. Make sure docker is installed locally.
3. Once installed, run ``make generate`` to generate all the code and mock client for FlyteAdmin Service aswell as the docs for it.

```bash
    make generate
```

4. To add new dependencies for documentation generation, modify ``doc-requirements.in`` and then

```bash
   make doc-requirements.txt
```

## Docs structure

The index.rst files for protos are kept in parallel folder structure under the docs folder.
All the proto definitions are within protos/flyteidl and there corresponding docs are kept in protos/docs

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

Each module in protos has same named module under the docs also.
eg : protos/flyteidl/core has same named doc structure placing it index and other documentation files in protos/docs/core


## Docs Generation

* If introducing a new module then follow the structure for core files in `generate_protos.sh` file which helps in generating the core documentation from its proto files.
```
     core_proto_files=`ls protos/flyteidl/core/*.proto |xargs`
     # Remove any currently generated file
     ls -d protos/docs/core/* | grep -v index.rst | xargs rm
     protoc --doc_out=protos/docs/core --doc_opt=restructuredtext,core.rst -I=protos `echo $core_proto_files`
```

* ``make generate`` would have already generated the modified rst files.

* ``make html`` Generate the sphinx documentation from the docs folder to use the modified rst for docs.


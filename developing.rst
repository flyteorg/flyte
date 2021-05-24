Contributing to Flyte IDL
=========================

This is one of the core repositories of Flyte and contains the Specification of
the Flyte Lanugage. The Specification is maintained using Googles fantastic
Protcol buffers library. The code and docs are auto-generated.

* [flyte.org](https://flyte.org)
* [Flyte Docs](http://flyte.readthedocs.org/)
* [FlyteIDL Docs](http://flyteidl.readthedocs.org)

Generate Code from protobuf
----------------------------
#. Make sure docker is installed locally.
#. Once installed run, ``make generate`` to generate all the code and mock
   client for FlyteAdmin Service

   .. prompt:: bash
    
     make generate

#. You might have to run, ``make download_tooling`` to install generator
   dependencies

   .. prompt:: bash

     make download_tooling

Generate Docs from protobuf
----------------------------

#. To add new dependencies for a doc, modify ``doc_requirements.in`` and then

   .. prompt:: bash

     make doc-requirements.txt

Docs structure
**************

The index.rst files for protos are kept in parallel folder structure under the docs folder.
All the proto definitions are within protos/flyteidl and there corresponding docs are kept in protos/docs

Each module in protos has same named module under the docs also.
eg : protos/flyteidl/core has same named doc structure placing it index and other documentation files in protos/docs/core


Docs Generation
***************

Relies on protoc plugin pseudomuto/protoc-gen-doc for generating the docs.
Currently pseudomuto/protoc-gen-doc is being used for generating the RST files which are placed in protos/docs and corresponding folder based on the module.
Until this code is checkedin https://github.com/pseudomuto/protoc-gen-doc/pull/440, use the branch to build the protoc-gen-doc executable
Follow the steps to build and use the protoc-gen-doc

#. This will install all the tooling including protoc-gen-doc plugin for protoc used for documentation

   .. prompt:: bash

      make download_tooling

#. The protoc-gen-doc will now be available for protoc.Following is an example from `generate_protos.sh` file which helps in generating the core documentation from its proto files

   .. prompt:: bash

     core_proto_files=`ls protos/flyteidl/core/*.proto |xargs`
     # Remove any currently generated file
     ls -d protos/docs/core/* | grep -v index.rst | xargs rm
     protoc --doc_out=protos/docs/core --doc_opt=restructuredtext,core.rst -I=protos `echo $core_proto_files`

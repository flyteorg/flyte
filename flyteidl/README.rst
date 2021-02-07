================
Flyte IDL
================
This is one of the core repositories of Flyte and contains the Specification of
the Flyte Lanugage. The Specification is maintained using Googles fantastic
Protcol buffers library. The code and docs are auto-generated.

 - [flyte.org](https://flyte.org)
 - [Flyte Docs](http://flyte.readthedocs.org/)
 - [FlyteIDL Docs](http://flyteidl.readthedocs.org)

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

#. To add new dependencies for a doc, modify ``doc_requirements.in`` and then

   .. prompt:: bash

   make doc-requirements.txt

Contributing to Flyte IDL
=========================

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


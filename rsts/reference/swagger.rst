.. _reference-swagger:

################################
Flyte API playground: Swagger
################################

Flyte services expose gRPC services for efficient/low latency communication across services as well as for external clients (flytectl, flyteconsole, flytekit Remote and flyte-cli,... etc.).

The service definitions are defined `here <https://github.com/flyteorg/flyteidl/tree/master/protos/flyteidl/service>`_. FlyteIdl also houses open API schema definitions for the exposed services.
You can find these here: `admin <https://github.com/flyteorg/flyteidl/blob/master/gen/pb-go/flyteidl/service/admin.swagger.json>`_, `auth <https://github.com/flyteorg/flyteidl/blob/master/gen/pb-go/flyteidl/service/auth.swagger.json>`_ and `identity <https://github.com/flyteorg/flyteidl/blob/master/gen/pb-go/flyteidl/service/identity.swagger.json>`_

For convenience, flyte deployments also ship with a `redocly/redoc <https://github.com/Redocly/redoc>`_ image container to expose swagger UI. By default, it's only deployed in sandbox deployments. To view that UI, you can: 

.. prompt:: bash $

   flytectl sandbox start

Once it's ready, you will get this message:

.. prompt:: bash $

   👨‍💻 Flyte is ready! Flyte UI is available at http://localhost:30081/console 🚀 🚀 🎉 

You can now visit http://localhost:30081/openapi to view the swagger UI for the service, documentation for the payload fields and send sample queries.

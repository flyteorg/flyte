.. _divedeep-domains:

Domains
=======

Domains provide an abstraction to isolate resources and feature configuration for different
deployment environments. 

For example: At Lyft, we develop and deploy Flyte workflows in development, staging, and production. We configure Flyte domains with those names, and specify lower resource limits on the development and staging domains than production domains. 

We also use ``Domains`` to disable Launchplans and Schedules from development and staging domains, since those features are typically meant for production deployments.

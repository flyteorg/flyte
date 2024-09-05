=====================
Configuration
=====================

.. currentmodule:: flytekit.configuration

Flytekit Configuration Sources
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

There are multiple ways to configure flytekit settings:

**Command Line Arguments**: This is the recommended way of setting configuration values for many cases.
For example, see `pyflyte package <pyflyte.html#pyflyte-package>`_ command.

**Python Config Object**: A :py:class:`~flytekit.configuration.Config` object can by used directly, e.g. when
initializing a :py:class:`~flytefit.remote.remote.FlyteRemote` object. See :doc:`here <../design/control_plane>` for examples on
how to specify a ``Config`` object.

**Environment Variables**: Users can specify these at compile time, but when your task is run, Flyte Propeller will
also set configuration to ensure correct interaction with the platform. The environment variables must be specified
with the format ``FLYTE_{SECTION}_{OPTION}``, all in upper case. For example, to specify the
:py:class:`PlatformConfig.endpoint <flytekit.configuration.PlatformConfig>` setting, the environment variable would
be ``FLYTE_PLATFORM_URL``.

.. note::

   Environment variables won't work for image configuration, which need to be specified with the
   `pyflyte package --image ... <pyflyte.html#cmdoption-pyflyte-package-i>`_ option or in a configuration
   file.

**YAML Format Configuration File**: A configuration file that contains settings for both
`flytectl <https://docs.flyte.org/en/latest/flytectl/overview.html>`__ and ``flytekit``. This is the recommended configuration
file format. Invoke the :ref:`flytectl config init <flytectl_config_init>` command to create a boilerplate
``~/.flyte/config.yaml`` file, and  ``flytectl --help`` to learn about all of the configuration yaml options.

.. dropdown:: See example ``config.yaml`` file
   :animate: fade-in-slide-down

   .. rli:: https://raw.githubusercontent.com/flyteorg/flytekit/master/tests/flytekit/unit/configuration/configs/sample.yaml
      :language: yaml
      :caption: config.yaml

**INI Format Configuration File**: A configuration file for ``flytekit``. By default, ``flytekit`` will look for a
file in two places:

1. First, a file named ``flytekit.config`` in the Python interpreter's working directory.
2. A file in ``~/.flyte/config`` in the home directory as detected by Python.

.. dropdown:: See example ``flytekit.config`` file
   :animate: fade-in-slide-down

   .. rli:: https://raw.githubusercontent.com/flyteorg/flytekit/master/tests/flytekit/unit/configuration/configs/images.config
      :language: ini
      :caption: flytekit.config

.. warning::

   The INI format configuration is considered a legacy configuration format. We recommend using the yaml format
   instead if you're using a configuration file.

How is configuration used?
^^^^^^^^^^^^^^^^^^^^^^^^^^

Configuration usage can roughly be bucketed into the following areas,

- **Compile-time settings**: these are settings like the default image and named images, where to look for Flyte code, etc.
- **Platform settings**: Where to find the Flyte backend (Admin DNS, whether to use SSL)
- **Registration Run-time settings**: these are things like the K8s service account to use, a specific S3/GCS bucket to write off-loaded data (dataframes and files) to, notifications, labels & annotations, etc.
- **Data access settings**: Is there a custom S3 endpoint in use? Backoff/retry behavior for accessing S3/GCS, key and password, etc.
- **Other settings** - Statsd configuration, which is a run-time applicable setting but is not necessarily relevant to the Flyte platform.

Configuration Objects
---------------------

The following objects are encapsulated in a parent object called ``Config``.

.. autosummary::
   :toctree: generated/
   :nosignatures:

   ~Config

.. _configuration-compile-time-settings:

Serialization Time Settings
^^^^^^^^^^^^^^^^^^^^^^^^^^^

These are serialization/compile-time settings that are used when using commands like
`pyflyte package <pyflyte.html#pyflyte-package>`_ or `pyflyte register <pyflyte.html#pyflyte-register>`_. These
configuration settings are typically passed in as flags to the above CLI commands.

The image configurations are typically either passed in via an `--image <pyflyte.html#cmdoption-pyflyte-package-i>`_ flag,
or can be specified in the ``yaml`` or ``ini`` configuration files (see examples above).

.. autosummary::
   :toctree: generated/
   :nosignatures:

   ~Image
   ~ImageConfig
   ~SerializationSettings
   ~FastSerializationSettings

.. _configuration-execution-time-settings:

Execution Time Settings
^^^^^^^^^^^^^^^^^^^^^^^

Users typically shouldn't be concerned with these configurations, as they are typically set by FlytePropeller or
FlyteAdmin. The configurations below are useful for authenticating to a Flyte backend, configuring data access
credentials, secrets, and statsd metrics.

.. autosummary::
   :toctree: generated/
   :nosignatures:

   ~PlatformConfig
   ~StatsConfig
   ~SecretsConfig
   ~S3Config
   ~GCSConfig
   ~DataConfig

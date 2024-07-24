import configparser
import datetime
import typing

from flytekit.configuration.file import ConfigEntry, ConfigFile, LegacyConfigEntry, YamlConfigEntry


class Images(object):
    @staticmethod
    def get_specified_images(cfg: typing.Optional[ConfigFile]) -> typing.Dict[str, str]:
        """
        This section should contain options, where the option name is the friendly name of the image and the corresponding
        value is actual FQN of the image. Example of how the section is structured
        [images]
        my_image1=docker.io/flyte:tag
        # Note that the tag is optional. If not specified it will be the default version identifier specified
        my_image2=docker.io/flyte

        :returns a dictionary of name: image<fqn+version> Version is optional
        """
        images: typing.Dict[str, str] = {}
        if cfg is None:
            return images

        if cfg.legacy_config:
            try:
                image_names = cfg.legacy_config.options("images")
            except configparser.NoSectionError:
                return {}
            if image_names:
                for i in image_names:
                    images[str(i)] = cfg.legacy_config.get("images", i)
            return images
        if cfg.yaml_config:
            return cfg.yaml_config.get("images", images)


class AWS(object):
    SECTION = "aws"
    S3_ENDPOINT = ConfigEntry(LegacyConfigEntry(SECTION, "endpoint"), YamlConfigEntry("storage.connection.endpoint"))
    S3_ACCESS_KEY_ID = ConfigEntry(
        LegacyConfigEntry(SECTION, "access_key_id"), YamlConfigEntry("storage.connection.access-key")
    )
    S3_SECRET_ACCESS_KEY = ConfigEntry(
        LegacyConfigEntry(SECTION, "secret_access_key"), YamlConfigEntry("storage.connection.secret-key")
    )
    ENABLE_DEBUG = ConfigEntry(LegacyConfigEntry(SECTION, "enable_debug", bool))
    RETRIES = ConfigEntry(LegacyConfigEntry(SECTION, "retries", int))
    BACKOFF_SECONDS = ConfigEntry(
        LegacyConfigEntry(SECTION, "backoff_seconds", datetime.timedelta),
        transform=lambda x: datetime.timedelta(seconds=int(x)),
    )


class GCP(object):
    SECTION = "gcp"
    GSUTIL_PARALLELISM = ConfigEntry(LegacyConfigEntry(SECTION, "gsutil_parallelism", bool))


class AZURE(object):
    SECTION = "azure"
    STORAGE_ACCOUNT_NAME = ConfigEntry(LegacyConfigEntry(SECTION, "storage_account_name"))
    STORAGE_ACCOUNT_KEY = ConfigEntry(LegacyConfigEntry(SECTION, "storage_account_key"))
    TENANT_ID = ConfigEntry(LegacyConfigEntry(SECTION, "tenant_id"))
    CLIENT_ID = ConfigEntry(LegacyConfigEntry(SECTION, "client_id"))
    CLIENT_SECRET = ConfigEntry(LegacyConfigEntry(SECTION, "client_secret"))


class Local(object):
    SECTION = "local"
    CACHE_ENABLED = ConfigEntry(LegacyConfigEntry(SECTION, "cache_enabled", bool))
    CACHE_OVERWRITE = ConfigEntry(LegacyConfigEntry(SECTION, "cache_overwrite", bool))


class Credentials(object):
    SECTION = "credentials"
    COMMAND = ConfigEntry(LegacyConfigEntry(SECTION, "command", list), YamlConfigEntry("admin.command", list))
    """
    This command is executed to return a token using an external process.
    """

    PROXY_COMMAND = ConfigEntry(
        LegacyConfigEntry(SECTION, "proxy_command", list), YamlConfigEntry("admin.proxyCommand", list)
    )
    """
    This command is executed to return a token for authorization with a proxy in front of Flyte using an external process.
    """

    CLIENT_ID = ConfigEntry(LegacyConfigEntry(SECTION, "client_id"), YamlConfigEntry("admin.clientId"))
    """
    This is the public identifier for the app which handles authorization for a Flyte deployment.
    More details here: https://www.oauth.com/oauth2-servers/client-registration/client-id-secret/.
    """

    CLIENT_CREDENTIALS_SECRET = ConfigEntry(LegacyConfigEntry(SECTION, "client_secret"))
    """
    Used for basic auth, which is automatically called during pyflyte. This will allow the Flyte engine to read the
    password directly from the environment variable. Note that this is less secure! Please only use this if mounting the
    secret as a file is impossible.
    """

    CLIENT_CREDENTIALS_SECRET_LOCATION = ConfigEntry(
        LegacyConfigEntry(SECTION, "client_secret_location"), YamlConfigEntry("admin.clientSecretLocation")
    )
    """
    Used for basic auth, which is automatically called during pyflyte. This will allow the Flyte engine to read the
    password from a mounted file.
    """

    CLIENT_CREDENTIALS_SECRET_ENV_VAR = ConfigEntry(
        LegacyConfigEntry(SECTION, "client_secret_env_var"), YamlConfigEntry("admin.clientSecretEnvVar")
    )
    """
    Used for basic auth, which is automatically called during pyflyte. This will allow the Flyte engine to read the
    password from a mounted environment variable.
    """

    SCOPES = ConfigEntry(LegacyConfigEntry(SECTION, "scopes", list), YamlConfigEntry("admin.scopes", list))
    """
    This setting can be used to manually pass in scopes into authenticator flows - eg.) for Auth0 compatibility
    """

    AUTH_MODE = ConfigEntry(LegacyConfigEntry(SECTION, "auth_mode"), YamlConfigEntry("admin.authType"))
    """
    The auth mode defines the behavior used to request and refresh credentials. The currently supported modes include:
    - 'standard' or 'Pkce': This uses the pkce-enhanced authorization code flow by opening a browser window to initiate
            credentials access.
    - "DeviceFlow": This uses the Device Authorization Flow
    - 'basic', 'client_credentials' or 'clientSecret': This uses symmetric key auth in which the end user enters a
            client id and a client secret and public key encryption is used to facilitate authentication.
    - None: No auth will be attempted.
    """


class Platform(object):
    SECTION = "platform"
    URL = ConfigEntry(
        LegacyConfigEntry(SECTION, "url"), YamlConfigEntry("admin.endpoint"), lambda x: x.replace("dns:///", "")
    )
    INSECURE = ConfigEntry(LegacyConfigEntry(SECTION, "insecure", bool), YamlConfigEntry("admin.insecure", bool))
    INSECURE_SKIP_VERIFY = ConfigEntry(
        LegacyConfigEntry(SECTION, "insecure_skip_verify", bool), YamlConfigEntry("admin.insecureSkipVerify", bool)
    )
    CONSOLE_ENDPOINT = ConfigEntry(LegacyConfigEntry(SECTION, "console_endpoint"), YamlConfigEntry("console.endpoint"))
    CA_CERT_FILE_PATH = ConfigEntry(
        LegacyConfigEntry(SECTION, "ca_cert_file_path"), YamlConfigEntry("admin.caCertFilePath")
    )
    HTTP_PROXY_URL = ConfigEntry(LegacyConfigEntry(SECTION, "http_proxy_url"), YamlConfigEntry("admin.httpProxyURL"))


class LocalSDK(object):
    SECTION = "sdk"
    WORKFLOW_PACKAGES = ConfigEntry(LegacyConfigEntry(SECTION, "workflow_packages", list))
    """
    This is a comma-delimited list of packages that SDK tools will use to discover entities for the purpose of registration
    and execution of entities.
    """

    LOCAL_SANDBOX = ConfigEntry(LegacyConfigEntry(SECTION, "local_sandbox"))
    """
    This is the path where SDK will place files during local executions and testing.  The SDK will not automatically
    clean up data in these directories.
    """

    LOGGING_LEVEL = ConfigEntry(LegacyConfigEntry(SECTION, "logging_level", int))
    """
    This is the default logging level for the Python logging library and will be set before user code runs.
    Note that this configuration is special in that it is a runtime setting, not a compile time setting.  This is the only
    runtime option in this file.

    TODO delete the one from internal config
    """


class Secrets(object):
    SECTION = "secrets"
    # Secrets management
    ENV_PREFIX = ConfigEntry(LegacyConfigEntry(SECTION, "env_prefix"))
    """
    This is the prefix that will be used to lookup for injected secrets at runtime. This can be overridden to using
    FLYTE_SECRETS_ENV_PREFIX variable
    """

    DEFAULT_DIR = ConfigEntry(LegacyConfigEntry(SECTION, "default_dir"))
    """
    This is the default directory that will be used to find secrets as individual files under. This can be overridden using
    FLYTE_SECRETS_DEFAULT_DIR.
    """

    FILE_PREFIX = ConfigEntry(LegacyConfigEntry(SECTION, "file_prefix"))
    """
    This is the prefix for the file in the default dir.
    """


class StatsD(object):
    SECTION = "secrets"
    # StatsD Config flags should ideally be controlled at the platform level and not through flytekit's config file.
    # They are meant to allow administrators to control certain behavior according to how the system is configured.

    HOST = ConfigEntry(LegacyConfigEntry(SECTION, "host"))
    PORT = ConfigEntry(LegacyConfigEntry(SECTION, "port", int))
    DISABLED = ConfigEntry(LegacyConfigEntry(SECTION, "disabled", bool))
    DISABLE_TAGS = ConfigEntry(LegacyConfigEntry(SECTION, "disable_tags", bool))

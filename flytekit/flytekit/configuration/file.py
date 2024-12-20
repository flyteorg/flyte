from __future__ import annotations

import configparser
import configparser as _configparser
import os
import pathlib
import typing
from dataclasses import dataclass
from os import getenv
from pathlib import Path

import yaml

from flytekit.exceptions import user as _user_exceptions
from flytekit.loggers import logger

# This is the env var that the flytectl sandbox instructions say to set
FLYTECTL_CONFIG_ENV_VAR = "FLYTECTL_CONFIG"
# This is an explicit override only to be used by pyflyte and takes precedence in get_config_file over the main
# environment variable.
# This env var should not be set by users
FLYTECTL_CONFIG_ENV_VAR_OVERRIDE = "_FLYTECTL_CONFIG_PYFLYTE_OVERRIDE"


def _exists(val: typing.Any) -> bool:
    """Check if a value is defined."""
    return isinstance(val, bool) or bool(val is not None and val)


@dataclass
class LegacyConfigEntry(object):
    """
    Creates a record for the config entry. contains
    Args:
        section: section the option should be found under
        option: the option str to lookup
        type_: Expected type of the value
    """

    section: str
    option: str
    type_: typing.Type = str

    def get_env_name(self) -> str:
        return f"FLYTE_{self.section.upper()}_{self.option.upper()}"

    def read_from_env(self, transform: typing.Optional[typing.Callable] = None) -> typing.Optional[typing.Any]:
        """
        Reads the config entry from environment variable, the structure of the env var is current
        ``FLYTE_{SECTION}_{OPTION}`` all upper cased. We will change this in the future.
        :return:
        """
        env = self.get_env_name()
        v = os.environ.get(env, None)
        if v is None:
            return None
        return transform(v) if transform else v

    def read_from_file(
        self, cfg: ConfigFile, transform: typing.Optional[typing.Callable] = None
    ) -> typing.Optional[typing.Any]:
        if not cfg:
            return None
        try:
            v = cfg.get(self)
            return transform(v) if transform else v
        except configparser.Error:
            pass
        return None


@dataclass
class YamlConfigEntry(object):
    """
    Creates a record for the config entry.
    Args:
        switch: dot-delimited string that should match flytectl args. Leaving it as dot-delimited instead of a list
          of strings because it's easier to maintain alignment with flytectl.
        config_value_type: Expected type of the value
    """

    switch: str
    config_value_type: typing.Type = str

    def read_from_file(
        self, cfg: ConfigFile, transform: typing.Optional[typing.Callable] = None
    ) -> typing.Optional[typing.Any]:
        if not cfg:
            return None
        try:
            v = cfg.get(self)
            if _exists(v):
                return transform(v) if transform else v
        except Exception:
            ...

        return None


def bool_transformer(config_val: typing.Any) -> bool:
    if type(config_val) is str:
        return True if config_val and config_val.lower() not in ["false", "0", "off", "no"] else False
    else:
        return config_val


def comma_list_transformer(config_val: typing.Any):
    if type(config_val) is str:
        return config_val.split(",")
    else:
        return config_val


def int_transformer(config_val: typing.Any):
    if type(config_val) is str:
        try:
            return int(config_val)
        except ValueError:
            logger.warning(f"Couldn't convert configuration setting {config_val} into {int}, leaving as is.")
    return config_val


@dataclass
class ConfigEntry(object):
    """
    A top level Config entry holder, that holds multiple different representations of the config.
    Legacy means the INI style config files. YAML support is for the flytectl config file, which is there by default
    when flytectl starts a sandbox
    """

    legacy: LegacyConfigEntry
    yaml_entry: typing.Optional[YamlConfigEntry] = None
    transform: typing.Optional[typing.Callable[[str], typing.Any]] = None

    legacy_default_transforms = {
        bool: bool_transformer,
        list: comma_list_transformer,
        int: int_transformer,
    }

    def __post_init__(self):
        if self.legacy:
            if not self.transform and self.legacy.type_ in ConfigEntry.legacy_default_transforms:
                self.transform = ConfigEntry.legacy_default_transforms[self.legacy.type_]

    def read(self, cfg: typing.Optional[ConfigFile] = None) -> typing.Optional[typing.Any]:
        """
        Reads the config Entry from the various sources in the following order,
        #. First try to read from the relevant environment variable,
        #. If missing, then try to read from the legacy config file, if one was parsed.
        #. If missing, then try to read from the yaml file.

        The constructor for ConfigFile currently does not allow specification of both the ini and yaml style formats.

        :param cfg:
        :return:
        """
        from_env = self.legacy.read_from_env(self.transform)
        if from_env is not None:
            return from_env
        if cfg and cfg.legacy_config:
            return self.legacy.read_from_file(cfg, self.transform)
        if cfg and cfg.yaml_config and self.yaml_entry:
            return self.yaml_entry.read_from_file(cfg, self.transform)

        return None


class ConfigFile(object):
    def __init__(self, location: str):
        """
        Load the config from this location
        """
        self._location = location
        if location.endswith("yaml") or location.endswith("yml"):
            self._legacy_config = None
            self._yaml_config = self._read_yaml_config(location)
        else:
            self._legacy_config = self._read_legacy_config(location)
            self._yaml_config = None

    @staticmethod
    def _read_yaml_config(location: str) -> typing.Optional[typing.Dict[str, typing.Any]]:
        with open(location, "r") as fh:
            try:
                yaml_contents = yaml.safe_load(fh)
                return yaml_contents
            except yaml.YAMLError as exc:
                logger.warning(f"Error {exc} reading yaml config file at {location}, ignoring...")
                return None

    def _read_legacy_config(self, location: str) -> _configparser.ConfigParser:
        c = _configparser.ConfigParser()
        c.read(self._location)
        if c.has_section("internal"):
            raise _user_exceptions.FlyteAssertion(
                "The config file '{}' cannot contain a section for internal " "only configurations.".format(location)
            )
        return c

    def _get_from_legacy(self, c: LegacyConfigEntry) -> typing.Any:
        if issubclass(c.type_, bool):
            return self._legacy_config.getboolean(c.section, c.option)

        if issubclass(c.type_, int):
            return self._legacy_config.getint(c.section, c.option)

        if issubclass(c.type_, list):
            v = self._legacy_config.get(c.section, c.option)
            return v.split(",")

        return self._legacy_config.get(c.section, c.option)

    def _get_from_yaml(self, c: YamlConfigEntry) -> typing.Any:
        keys = c.switch.split(".")  # flytectl switches are dot delimited
        d = self.yaml_config
        try:
            for k in keys:
                d = d[k]
            return d
        except KeyError:
            logger.debug(f"Switch {c.switch} could not be found in yaml config")
            return None

    def get(self, c: typing.Union[LegacyConfigEntry, YamlConfigEntry]) -> typing.Any:
        if isinstance(c, LegacyConfigEntry):
            return self._get_from_legacy(c)
        if isinstance(c, YamlConfigEntry) and self.yaml_config:
            return self._get_from_yaml(c)
        raise NotImplementedError("Support for other config types besides .ini / .config files not yet supported")

    @property
    def legacy_config(self) -> _configparser.ConfigParser:
        return self._legacy_config

    @property
    def yaml_config(self) -> typing.Dict[str, typing.Any]:
        return self._yaml_config


def get_config_file(c: typing.Union[str, ConfigFile, None]) -> typing.Optional[ConfigFile]:
    """
    Checks if the given argument is a file or a configFile and returns a loaded configFile else returns None
    """
    if c is None:
        # Pyflyte override env var takes highest precedence
        # Env var takes second highest precedence
        flytectl_path_from_env = getenv(FLYTECTL_CONFIG_ENV_VAR_OVERRIDE, getenv(FLYTECTL_CONFIG_ENV_VAR, None))
        if flytectl_path_from_env:
            flytectl_path = Path(flytectl_path_from_env)
            if flytectl_path.exists():
                logger.info(f"Using flytectl/YAML config {flytectl_path.absolute()}")
                return ConfigFile(str(flytectl_path.absolute()))
            else:
                logger.warning(f"flytectl config file {flytectl_path.absolute()} does not exist, ignoring...")

        # See if there's a config file in the current directory where Python is being run from
        current_location_config = Path("flytekit.config")
        if current_location_config.exists():
            logger.info(f"Using configuration from Python process root {current_location_config.absolute()}")
            return ConfigFile(str(current_location_config.absolute()))

        # If not, see if there's a config in the user's home directory
        home_dir_config = Path(Path.home(), ".flyte", "config")  # _default_config_file_name in main.py
        if home_dir_config.exists():
            logger.info(f"Using configuration from home directory {home_dir_config.absolute()}")
            return ConfigFile(str(home_dir_config.absolute()))

        # If not see if there's something in the default home directory location
        flytectl_path = Path(Path.home(), ".flyte", "config.yaml")
        if flytectl_path.exists():
            logger.info(f"Using flytectl/YAML config {flytectl_path.absolute()}")
            return ConfigFile(str(flytectl_path.absolute()))

        # If not, then return None and let caller handle
        return None
    if isinstance(c, str):
        logger.debug(f"Using specified config file at {c}")
        return ConfigFile(c)
    return c


def set_if_exists(d: dict, k: str, v: typing.Any) -> dict:
    """
    Given a dict ``d`` sets the key ``k`` with value of config ``v``, if the config value ``v`` is set
    and return the updated dictionary.

    .. note::

        The input dictionary ``d`` will be mutated.
    """
    if _exists(v):
        d[k] = v
    return d


def read_file_if_exists(filename: typing.Optional[str], encoding=None) -> typing.Optional[str]:
    """
    Reads the contents of the file if passed a path. Otherwise, returns None.

    :param filename: The file path to load
    :param encoding: The encoding to use when reading the file.
    :return: The contents of the file as a string or None.
    """
    if not filename:
        return None

    filename = pathlib.Path(filename)
    logger.debug(f"Reading file contents from [{filename}] with current directory [{os.getcwd()}].")
    return filename.read_text(encoding=encoding)

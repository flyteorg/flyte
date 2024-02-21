# -*- coding: utf-8 -*-
#
import re
import sys

import statsd

from flytekit.configuration import StatsConfig

RESERVED_TAG_WORDS = frozenset(
    ["asg", "az", "backend", "canary", "host", "period", "region", "shard", "window", "source"]
)

# TODO should this be a whitelist instead?
FORBIDDEN_TAG_VALUE_CHARACTERS = re.compile("[|.:]")

_stats_client = None


class ScopeableStatsProxy(object):
    """
    A Proxy object for an underlying statsd client.
    Adds a new call, scope(prefix), which returns a new proxy to the same
    client which will prefix all calls to underlying methods with the scoped prefix:
    new_client = client.get_stats('a')
    new_client.incr('b') # Metric name = a.b
    This can be nested:
    newer_client = new_client.get_stats('subsystem')
    newer_client.incr('bad') # Metric name = a.subsystem.bad
    """

    # List of functions we will proxy and prefix the first string argument on
    EXTENDABLE_FUNC = ["incr", "decr", "timing", "timer", "gauge", "set"]

    def __init__(self, client, prefix=None):
        self._client = client
        self._scope_prefix = prefix
        for extendable_func in self.EXTENDABLE_FUNC:
            base_func = getattr(self._client, extendable_func)
            if base_func:
                setattr(self, extendable_func, self._create_wrapped_function(base_func))

    def _create_wrapped_function(self, base_func):
        if self._scope_prefix:

            def name_wrap(stat, *args, **kwargs):
                tags = kwargs.pop("tags", {})
                if kwargs.pop("per_host", False):
                    tags["_f"] = "i"

                if bool(tags):
                    stat = self._serialize_tags(stat, tags)
                return base_func(self._p_with_prefix(stat), *args, **kwargs)

        else:

            def name_wrap(stat, *args, **kwargs):
                tags = kwargs.pop("tags", {})
                if kwargs.pop("per_host", False):
                    tags["_f"] = "i"

                if bool(tags):
                    stat = self._serialize_tags(stat, tags)
                return base_func(stat, *args, **kwargs)

        return name_wrap

    def get_stats(self, name):
        if not self._scope_prefix or self._scope_prefix == "":
            prefix = name
        else:
            prefix = self._scope_prefix + "." + name

        return ScopeableStatsProxy(self._client, prefix)

    def pipeline(self):
        return ScopeableStatsProxy(self._client.pipeline(), self._scope_prefix)

    def _p_with_prefix(self, name):
        if name is None:
            return name
        return self._scope_prefix + "." + name

    def _is_ascii(self, name):
        if sys.version_info >= (3, 7):
            return name.isascii()
        try:
            return name and name.encode("ascii")
        except UnicodeEncodeError:
            return False

    def _serialize_tags(self, metric, tags=None):
        stat = metric
        if tags:
            for key in sorted(tags):
                try:
                    # Python 2.7, will throw UnicodeEncodeError
                    # Python 3.4+, will need ascii check,
                    # pystatsd  3.2.1 can handle ascii only (.encode with "ignore" can be lossy)
                    key = str(key)
                    # if the tags fail sanity check we will not serialize the tags, and simply return the metric.
                    if not self._is_ascii(key):
                        return stat
                    tag = FORBIDDEN_TAG_VALUE_CHARACTERS.sub("_", str(tags[key]))
                    if tag != "":
                        metric += ".__{0}={1}".format(key, tag)
                except UnicodeEncodeError:
                    # TODO: log warning here and possibly fail the entire request?
                    return stat

        return metric

    def __hasattr__(self, name):
        return hasattr(self._client, name)

    def __getattr__(self, name):
        return getattr(self._client, name)

    def __enter__(self):
        return ScopeableStatsProxy(self._client.__enter__(), self._scope_prefix)

    def __exit__(self, exc_type, exc_value, traceback):
        self._client.__exit__(exc_type, exc_value, traceback)


class StatsClientProxy(ScopeableStatsProxy):
    @property
    def _prefix(self):
        return self._scope_prefix


def _get_stats_client(cfg: StatsConfig):
    global _stats_client
    if cfg.disabled is True:
        _stats_client = DummyStatsClient()
    if _stats_client is None:
        _stats_client = statsd.StatsClient(cfg.host, cfg.port)
    return _stats_client


def get_base_stats(cfg: StatsConfig, prefix: str):
    return StatsClientProxy(_get_stats_client(cfg), prefix=prefix)


def get_stats(cfg: StatsConfig, prefix: str):
    return get_base_stats(cfg, prefix)


class DummyStatsClient(statsd.StatsClient):
    """A dummy client for statsd."""

    def __init__(self, host="localhost", port=8125, prefix=None, maxudpsize=512, ipv6=False):
        super().__init__(host, port, prefix, maxudpsize, ipv6)

    def _send(self, data):
        pass

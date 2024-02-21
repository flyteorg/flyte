from typing import Dict

from flytekit.configuration import StatsConfig
from flytekit.interfaces.stats import client as _stats_client


class TaggableStats(_stats_client.ScopeableStatsProxy):
    # List of functions we will proxy and prefix the first string argument on
    EXTENDABLE_FUNC = ["incr", "decr", "timing", "timer", "gauge", "set"]

    def __init__(self, client, full_prefix, cfg: StatsConfig, prefix=None, tags=None):
        super(TaggableStats, self).__init__(client, prefix=prefix)
        self._tags = tags if tags else {}
        self._full_prefix = full_prefix
        self._cfg = cfg

    def _create_wrapped_function(self, base_func):
        if self._scope_prefix:

            def name_wrap(stat, *args, **kwargs):
                tags = kwargs.pop("tags", {})
                tags.update(self._tags)
                if kwargs.pop("per_host", False):
                    tags["_f"] = "i"

                if bool(tags) and not self._cfg.disabled_tags:
                    stat = self._serialize_tags(stat, tags)
                return base_func(self._p_with_prefix(stat), *args, **kwargs)

        else:

            def name_wrap(stat, *args, **kwargs):
                tags = kwargs.pop("tags", {})
                tags.update(self._tags)
                if kwargs.pop("per_host", False):
                    tags["_f"] = "i"

                if bool(tags) and not self._cfg.disabled_tags:
                    stat = self._serialize_tags(stat, tags)
                return base_func(stat, *args, **kwargs)

        return name_wrap

    def clear_tags(self):
        self._tags = {}

    def extend_tags(self, tags):
        self._tags.update(tags)

    def pipeline(self):
        return TaggableStats(
            self._client.pipeline(),
            self._full_prefix,
            cfg=self._cfg,
            prefix=self._scope_prefix,
            tags=dict(self._tags),
        )

    def __enter__(self):
        return TaggableStats(
            self._client.__enter__(),
            self._full_prefix,
            cfg=self._cfg,
            prefix=self._scope_prefix,
            tags=dict(self._tags),
        )

    def get_stats(self, name, copy_tags=True):
        if not self._scope_prefix or self._scope_prefix == "":
            prefix = name
        else:
            prefix = self._scope_prefix + "." + name

        if self._full_prefix:
            full_prefix = self._full_prefix + "." + prefix
        else:
            full_prefix = prefix

        tags = dict(self._tags) if copy_tags else None
        return TaggableStats(self._client, full_prefix, prefix=prefix, tags=tags)

    @property
    def full_prefix(self):
        return self._full_prefix


def get_stats(cfg: StatsConfig, prefix: str, tags: Dict[str, str] = None) -> TaggableStats:
    """
    :rtype: TaggableStats
    """

    # If tagging is disabled, do not pass tags to the constructor.
    if cfg.disabled_tags:
        tags = None

    return TaggableStats(_stats_client.get_base_stats(cfg, prefix.lower()), prefix.lower(), cfg=cfg, tags=tags)

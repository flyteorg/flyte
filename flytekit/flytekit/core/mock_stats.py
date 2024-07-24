import datetime as _datetime

from flytekit.loggers import logger


class MockStats(object):
    def __init__(self, scope="", tags=None):
        """
        Initializes a new mock stats object
        :param Text scope:
        :param dict[Text, Text] tags:
        """
        self.scope = scope
        self.tags = tags
        self._records = {}
        self._records_tags = {}

    def incr(self, metric, count=1, tags=None, **kwargs):
        full_name = self.scope + "." + metric
        self._records[full_name] = self._records.get(full_name, 0) + count
        self._records_tags[full_name] = tags or {}

    def decr(self, metric, count=1, tags=None, **kwargs):
        full_name = self.scope + "." + metric
        self._records[full_name] = self._records.get(full_name, 0) - count
        self._records_tags[full_name] = tags or {}

    def timing(self, metric):
        logger.warning("mock timing isn't implemented yet.")

    def timer(self, metric, tags=None, **kwargs):
        return _Timer(self, metric, tags=tags or {})

    def gauge(self, metric, value, tags=None, **kwargs):
        full_name = self.scope + "." + metric
        self._records[full_name] = value
        self._records_tags[full_name] = tags or {}

    def current_value(self, metric):
        full_name = self.scope + "." + metric
        return self._records.get(full_name, None)

    def current_tags(self, metric):
        full_name = self.scope + "." + metric
        return self._records_tags.get(full_name, None)


class _Timer(object):
    def __init__(self, mock_stats, metric, tags):
        """
        :param MockStats mock_stats:
        :param Text metric:
        """
        self._mock_stats = mock_stats
        self._metric = metric
        self._timer = None
        self._tags = tags

    def __enter__(self):
        self._timer = _datetime.datetime.utcnow()

    def __exit__(self, exc_type, exc_val, exc_tb):
        self._mock_stats.gauge(self._metric, _datetime.datetime.utcnow() - self._timer, tags=self._tags)
        self._timer = None

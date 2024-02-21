"""
This is a deprecated module. Model files for plugins should go alongside the microlib.
See ``plugins/flytekit-kf-pytorch/flytekitplugins/kfpytorch/models.py`` as an example.
"""

from flyteidl.plugins import qubole_pb2 as _qubole

from flytekit.models import common as _common


class HiveQuery(_common.FlyteIdlEntity):
    def __init__(self, query, timeout_sec, retry_count):
        """
        Initializes a new HiveQuery.

        :param Text query: The query string.
        :param int timeout_sec:
        :param int retry_count:

        """
        self._query = query
        self._timeout_sec = timeout_sec
        self._retry_count = retry_count

    @property
    def query(self):
        """
        The query string.
        :rtype: str
        """
        return self._query

    @property
    def timeout_sec(self):
        """
        :rtype: int
        """
        return self._timeout_sec

    @property
    def retry_count(self):
        """
        :rtype: int
        """
        return self._retry_count

    def to_flyte_idl(self):
        """
        :rtype: _qubole.HiveQuery
        """
        return _qubole.HiveQuery(query=self.query, timeout_sec=self.timeout_sec, retryCount=self.retry_count)

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param _qubole.HiveQuery pb2_object:
        :return: HiveQuery
        """
        return cls(
            query=pb2_object.query,
            timeout_sec=pb2_object.timeout_sec,
            retry_count=pb2_object.retryCount,
        )


class HiveQueryCollection(_common.FlyteIdlEntity):
    def __init__(self, queries):
        """
        Initializes a new HiveQueryCollection.

        :param list[HiveQuery] queries: Queries to execute.
        """
        self._queries = queries

    @property
    def queries(self):
        """
        :rtype: list[HiveQuery]
        """
        return self._queries

    def to_flyte_idl(self):
        """
        :rtype: _qubole.HiveQueryCollection
        """
        return _qubole.HiveQueryCollection(
            queries=[query.to_flyte_idl() for query in self.queries] if self.queries else None
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param _qubole.HiveQuery pb2_object:
        :rtype: HiveQueryCollection
        """
        return cls(queries=[HiveQuery.from_flyte_idl(query) for query in pb2_object.queries])


class QuboleHiveJob(_common.FlyteIdlEntity):
    def __init__(self, query, cluster_label, tags, query_collection=None):
        """
        Initializes a HiveJob.

        :param HiveQuery query: Single query to execute
        :param Text cluster_label: The qubole cluster label to execute the query on
        :param list[Text] tags: User tags for the queries
        :param HiveQueryCollection query_collection: Deprecated Queries to execute.
        """
        self._query = query
        self._cluster_label = cluster_label
        self._tags = tags
        self._query_collection = query_collection

    @property
    def query_collection(self):
        """
        The queries to be executed
        :rtype: HiveQueryCollection
        """
        return self._query_collection

    @property
    def query(self):
        """
        The query to be executed
        :rtype: HiveQuery
        """
        return self._query

    @property
    def cluster_label(self):
        """
        The cluster label where the query should be executed
        :rtype: Text
        """
        return self._cluster_label

    @property
    def tags(self):
        """
        User tags for the queries
        :rtype: list[Text]
        """
        return self._tags

    def to_flyte_idl(self):
        """
        :rtype: _qubole.QuboleHiveJob
        """
        return _qubole.QuboleHiveJob(
            query_collection=self._query_collection.to_flyte_idl() if self._query_collection else None,
            query=self._query.to_flyte_idl() if self._query else None,
            cluster_label=self._cluster_label,
            tags=self._tags,
        )

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param _qubole.QuboleHiveJob p:
        :rtype: QuboleHiveJob
        """
        return cls(
            query_collection=HiveQueryCollection.from_flyte_idl(p.query_collection)
            if p.HasField("query_collection")
            else None,
            query=HiveQuery.from_flyte_idl(p.query) if p.HasField("query") else None,
            cluster_label=p.cluster_label,
            tags=p.tags,
        )

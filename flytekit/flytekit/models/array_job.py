import json as _json

from flyteidl.plugins import array_job_pb2 as _array_job
from google.protobuf import json_format as _json_format

from flytekit.models import common as _common


class ArrayJob(_common.FlyteCustomIdlEntity):
    def __init__(self, parallelism=None, size=None, min_successes=None, min_success_ratio=None):
        """
        Initializes a new ArrayJob.
        :param int parallelism: Defines the minimum number of instances to bring up concurrently at any given point.
        :param int size: Defines the number of instances to launch at most. This number should match the size of
            the input if the job requires processing of all input data. This has to be a positive number.
        :param int min_successes: An absolute number of the minimum number of successful completions of subtasks. As
            soon as this criteria is met, the array job will be marked as successful and outputs will be computed.
        :param float min_success_ratio: Determines the minimum fraction of total jobs which can complete successfully
            before terminating the job and marking it successful.
        """
        if min_successes and min_success_ratio:
            raise ValueError("Only one of min_successes or min_success_ratio can be set")
        self._parallelism = parallelism
        self._size = size
        self._min_successes = min_successes
        self._min_success_ratio = min_success_ratio

    @property
    def parallelism(self):
        """
        Defines the minimum number of instances to bring up concurrently at any given point.

        :rtype: int
        """
        return self._parallelism

    @property
    def size(self):
        """
         Defines the number of instances to launch at most. This number should match the size of the input if the job
         requires processing of all input data. This has to be a positive number.

        :rtype: int
        """
        return self._size

    @size.setter
    def size(self, value):
        self._size = value

    @property
    def min_successes(self):
        """
        An absolute number of the minimum number of successful completions of subtasks. As soon as this criteria is met,
            the array job will be marked as successful and outputs will be computed.

        :rtype: int
        """
        return self._min_successes

    @property
    def min_success_ratio(self):
        return self._min_success_ratio

    @min_successes.setter
    def min_successes(self, value):
        self._min_successes = value

    def to_dict(self):
        """
        :rtype: dict[T, Text]
        """
        array_job = None
        if self.min_successes is not None:
            array_job = _array_job.ArrayJob(
                parallelism=self.parallelism,
                size=self.size,
                min_successes=self.min_successes,
            )
        elif self.min_success_ratio is not None:
            array_job = _array_job.ArrayJob(
                parallelism=self.parallelism,
                size=self.size,
                min_success_ratio=self.min_success_ratio,
            )

        return _json_format.MessageToDict(array_job)

    @classmethod
    def from_dict(cls, idl_dict):
        """
        :param dict[T, Text] idl_dict:
        :rtype: ArrayJob
        """
        pb2_object = _json_format.Parse(_json.dumps(idl_dict), _array_job.ArrayJob())

        if pb2_object.HasField("min_successes"):
            return cls(
                parallelism=pb2_object.parallelism,
                size=pb2_object.size,
                min_successes=pb2_object.min_successes,
            )
        else:
            return cls(
                parallelism=pb2_object.parallelism,
                size=pb2_object.size,
                min_success_ratio=pb2_object.min_success_ratio,
            )

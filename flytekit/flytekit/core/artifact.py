from __future__ import annotations

import datetime
import typing
from datetime import timedelta
from typing import Optional, Union

from flyteidl.core import artifact_id_pb2 as art_id
from google.protobuf.timestamp_pb2 import Timestamp

TIME_PARTITION_KWARG = "time_partition"


class InputsBase(object):
    """
    A class to provide better partition semantics
    Used for invoking an Artifact to bind partition keys to input values.
    If there's a good reason to use a metaclass in the future we can, but a simple instance suffices for now
    """

    def __getattr__(self, name: str) -> art_id.InputBindingData:
        return art_id.InputBindingData(var=name)


Inputs = InputsBase()


class ArtifactIDSpecification(object):
    """
    This is a special object that helps specify how Artifacts are to be created. See the comment in the
    call function of the main Artifact class. Also see the handling code in transform_variable_map for more
    information. There's a limited set of information that we ultimately need in a TypedInterface, so it
    doesn't make sense to carry the full Artifact object around. This object should be sufficient, despite
    having a pointer to the main artifact.
    """

    def __init__(self, a: Artifact):
        self.artifact = a
        self.partitions: Optional[Partitions] = None
        self.time_partition: Optional[TimePartition] = None

    # todo: add time partition arg hint
    def __call__(self, *args, **kwargs):
        return self.bind_partitions(*args, **kwargs)

    def bind_partitions(self, *args, **kwargs) -> ArtifactIDSpecification:
        # See the parallel function in the main Artifact class for more information.
        if len(args) > 0:
            raise ValueError("Cannot set partition values by position")

        if TIME_PARTITION_KWARG in kwargs:
            if not self.artifact.time_partitioned:
                raise ValueError("Cannot bind time partition to non-time partitioned artifact")
            p = kwargs[TIME_PARTITION_KWARG]
            if isinstance(p, datetime.datetime):
                t = Timestamp()
                t.FromDatetime(p)
                self.time_partition = TimePartition(value=art_id.LabelValue(time_value=t))
            elif isinstance(p, art_id.InputBindingData):
                self.time_partition = TimePartition(value=art_id.LabelValue(input_binding=p))
            else:
                raise ValueError(f"Time partition needs to be input binding data or static string, not {p}")
            # Given the context, shouldn't need to set further reference_artifacts.

            del kwargs[TIME_PARTITION_KWARG]

        if len(kwargs) > 0:
            p = Partitions(None)
            # k is the partition key, v should be static, or an input to the task or workflow
            for k, v in kwargs.items():
                if not self.artifact.partition_keys or k not in self.artifact.partition_keys:
                    raise ValueError(f"Partition key {k} not found in {self.artifact.partition_keys}")
                if isinstance(v, art_id.InputBindingData):
                    p.partitions[k] = Partition(art_id.LabelValue(input_binding=v), name=k)
                elif isinstance(v, str):
                    p.partitions[k] = Partition(art_id.LabelValue(static_value=v), name=k)
                else:
                    raise ValueError(f"Partition key {k} needs to be input binding data or static string, not {v}")
            # Given the context, shouldn't need to set further reference_artifacts.
            self.partitions = p

        return self

    def to_partial_artifact_id(self) -> art_id.ArtifactID:
        # This function should only be called by transform_variable_map
        artifact_id = self.artifact.to_id_idl()
        # Use the partitions from this object, but replacement is not allowed by protobuf, so generate new object
        p = Serializer.partitions_to_idl(self.partitions)
        tp = None
        if self.artifact.time_partitioned:
            if not self.time_partition:
                raise ValueError(
                    f"Artifact {artifact_id.artifact_key} requires a time partition, but it hasn't been bound."
                )
            tp = self.time_partition.to_flyte_idl()

        if self.artifact.partition_keys:
            required = len(self.artifact.partition_keys)
            fulfilled = len(p.value) if p else 0
            if required != fulfilled:
                raise ValueError(
                    f"Artifact {artifact_id.artifact_key} requires {required} partitions, but only {fulfilled} are "
                    f"bound."
                )
        artifact_id = art_id.ArtifactID(
            artifact_key=artifact_id.artifact_key,
            partitions=p,
            time_partition=tp,
            version=artifact_id.version,  # this should almost never be set since setting it
            # hardcodes the query to one version
        )
        return artifact_id


class ArtifactQuery(object):
    def __init__(
        self,
        artifact: Artifact,
        name: str,
        project: Optional[str] = None,
        domain: Optional[str] = None,
        time_partition: Optional[TimePartition] = None,
        partitions: Optional[Partitions] = None,
        tag: Optional[str] = None,
    ):
        if not name:
            raise ValueError("Cannot create query without name")

        # So normally, if you just do MyData.query(partitions="region": Inputs.region), it will just
        # use the input value to fill in the partition. But if you do
        #   MyData.query(region=OtherArtifact.partitions.region)
        # then you now have a dependency on the other artifact. This list keeps track of all the other Artifacts you've
        # referenced.
        self.artifact = artifact
        bindings: typing.List[Artifact] = []
        if time_partition:
            if time_partition.reference_artifact and time_partition.reference_artifact is not artifact:
                bindings.append(time_partition.reference_artifact)
        if partitions and partitions.partitions:
            for k, v in partitions.partitions.items():
                if v.reference_artifact and v.reference_artifact is not artifact:
                    bindings.append(v.reference_artifact)

        self.name = name
        self.project = project
        self.domain = domain
        self.time_partition = time_partition
        self.partitions = partitions
        self.tag = tag
        self.bindings = bindings

    def to_flyte_idl(
        self,
        **kwargs,
    ) -> art_id.ArtifactQuery:
        return Serializer.artifact_query_to_idl(self, **kwargs)


class TimePartition(object):
    def __init__(
        self,
        value: Union[art_id.LabelValue, art_id.InputBindingData, str, datetime.datetime, None],
        op: Optional[str] = None,
        other: Optional[timedelta] = None,
    ):
        if isinstance(value, str):
            raise ValueError(f"value to a time partition shouldn't be a str {value}")
        elif isinstance(value, datetime.datetime):
            t = Timestamp()
            t.FromDatetime(value)
            value = art_id.LabelValue(time_value=t)
        elif isinstance(value, art_id.InputBindingData):
            value = art_id.LabelValue(input_binding=value)
        # else should already be a LabelValue or None
        self.value: art_id.LabelValue = value
        self.op = op
        self.other = other
        self.reference_artifact: Optional[Artifact] = None

    def __add__(self, other: timedelta) -> TimePartition:
        tp = TimePartition(self.value, op="+", other=other)
        tp.reference_artifact = self.reference_artifact
        return tp

    def __sub__(self, other: timedelta) -> TimePartition:
        tp = TimePartition(self.value, op="-", other=other)
        tp.reference_artifact = self.reference_artifact
        return tp

    def to_flyte_idl(self, **kwargs) -> Optional[art_id.TimePartition]:
        return Serializer.time_partition_to_idl(self, **kwargs)


class Partition(object):
    def __init__(self, value: Optional[art_id.LabelValue], name: str):
        self.name = name
        self.value = value
        self.reference_artifact: Optional[Artifact] = None


class Partitions(object):
    def __init__(self, partitions: Optional[typing.Mapping[str, Union[str, art_id.InputBindingData, Partition]]]):
        self._partitions = {}
        if partitions:
            for k, v in partitions.items():
                if isinstance(v, Partition):
                    self._partitions[k] = v
                elif isinstance(v, art_id.InputBindingData):
                    self._partitions[k] = Partition(art_id.LabelValue(input_binding=v), name=k)
                else:
                    self._partitions[k] = Partition(art_id.LabelValue(static_value=v), name=k)
        self.reference_artifact: Optional[Artifact] = None

    @property
    def partitions(self) -> Optional[typing.Dict[str, Partition]]:
        return self._partitions

    def set_reference_artifact(self, artifact: Artifact):
        self.reference_artifact = artifact
        if self.partitions:
            for p in self.partitions.values():
                p.reference_artifact = artifact

    def __getattr__(self, item):
        if self.partitions and item in self.partitions:
            return self.partitions[item]
        raise AttributeError(f"Partition {item} not found in {self}")

    def to_flyte_idl(self, **kwargs) -> Optional[art_id.Partitions]:
        return Serializer.partitions_to_idl(self, **kwargs)


class Artifact(object):
    """
    An Artifact is effectively just a metadata layer on top of data that exists in Flyte. Most data of interest
    will be the output of tasks and workflows. The other category is user uploads.

    This Python class has limited purpose, as a way for users to specify that tasks/workflows create Artifacts
    and the manner (i.e. name, partitions) in which they are created.

    Control creation parameters at task/workflow execution time ::

        @task
        def t1() -> Annotated[nn.Module, Artifact(name="my.artifact.name")]:
            ...
    """

    def __init__(
        self,
        project: Optional[str] = None,
        domain: Optional[str] = None,
        name: Optional[str] = None,
        version: Optional[str] = None,
        time_partitioned: bool = False,
        time_partition: Optional[TimePartition] = None,
        partition_keys: Optional[typing.List[str]] = None,
        partitions: Optional[Union[Partitions, typing.Dict[str, str]]] = None,
    ):
        """
        :param project: Should not be directly user provided, the project/domain will come from the project/domain of
           the execution that produced the output. These values will be filled in automatically when retrieving however.
        :param domain: See above.
        :param name: The name of the Artifact. This should be user provided.
        :param version: Version of the Artifact, typically the execution ID, plus some additional entropy.
           Not user provided.
        :param time_partitioned: Whether or not this Artifact will have a time partition.
        :param partition_keys: This is a list of keys that will be used to partition the Artifact. These are not the
           values. Values are set via a () on the artifact and will end up in the partition_values field.
        :param partitions: This is a dictionary of partition keys to values.
        """
        if not name:
            raise ValueError("Can't instantiate an Artifact without a name.")
        self.project = project
        self.domain = domain
        self.name = name
        self.version = version
        self.time_partitioned = time_partitioned
        self._time_partition = None
        if time_partition:
            self._time_partition = time_partition
            self._time_partition.reference_artifact = self
        self.partition_keys = partition_keys
        self._partitions: Optional[Partitions] = None
        if partitions:
            if isinstance(partitions, dict):
                self._partitions = Partitions(partitions)
                self.partition_keys = list(partitions.keys())
            elif isinstance(partitions, Partitions):
                self._partitions = partitions
                if not partitions.partitions:
                    raise ValueError("Partitions must be non-empty")
                self.partition_keys = list(partitions.partitions.keys())
            else:
                raise ValueError(f"Partitions must be a dict or Partitions object, not {type(partitions)}")
            self._partitions.set_reference_artifact(self)
        if not partitions and partition_keys:
            # this should be the only time where we create Partition objects with None
            p = {k: Partition(None, name=k) for k in partition_keys}
            self._partitions = Partitions(p)
            self._partitions.set_reference_artifact(self)

    def __call__(self, *args, **kwargs) -> ArtifactIDSpecification:
        """
        This __call__ should only ever happen in the context of a task or workflow's output, to be
        used in an Annotated[] call. The other styles will go through different call functions.
        """
        # Can't guarantee the order in which time/non-time partitions are bound so create the helper
        # object and invoke the function there.
        partial_id = ArtifactIDSpecification(self)
        return partial_id.bind_partitions(*args, **kwargs)

    @property
    def partitions(self) -> Optional[Partitions]:
        return self._partitions

    @property
    def time_partition(self) -> TimePartition:
        if not self.time_partitioned:
            raise ValueError(f"Artifact {self.name} is not time partitioned")
        if not self._time_partition and self.time_partitioned:
            self._time_partition = TimePartition(None)
            self._time_partition.reference_artifact = self
        return self._time_partition

    def __str__(self):
        tp_str = f"  time partition={self.time_partition}\n" if self.time_partitioned else ""
        return (
            f"Artifact: project={self.project}, domain={self.domain}, name={self.name}, version={self.version}\n"
            f"  name={self.name}\n"
            f"  partitions={self.partitions}\n"
            f"{tp_str}"
        )

    def __repr__(self):
        return self.__str__()

    def query(
        self,
        project: Optional[str] = None,
        domain: Optional[str] = None,
        time_partition: Optional[Union[datetime.datetime, TimePartition, art_id.InputBindingData]] = None,
        partitions: Optional[Union[typing.Dict[str, str], Partitions]] = None,
        **kwargs,
    ) -> ArtifactQuery:
        if self.partition_keys:
            fn_args = {"project", "domain", "time_partition", "partitions", "tag"}
            k = set(self.partition_keys)
            if len(fn_args & k) > 0:
                raise ValueError(
                    f"There are conflicting partition key names {fn_args ^ k}, please rename"
                    f" use a partitions object"
                )
        if partitions and kwargs:
            raise ValueError("Please either specify kwargs or a partitions object not both")

        p_obj: Optional[Partitions] = None
        if kwargs:
            p_obj = Partitions(kwargs)
            p_obj.reference_artifact = self  # only set top level

        if partitions and isinstance(partitions, dict):
            p_obj = Partitions(partitions)
            p_obj.reference_artifact = self  # only set top level

        tp = None
        if time_partition:
            if isinstance(time_partition, TimePartition):
                tp = time_partition
            else:
                tp = TimePartition(time_partition)
                tp.reference_artifact = self

        tp = tp or (self.time_partition if self.time_partitioned else None)

        aq = ArtifactQuery(
            artifact=self,
            name=self.name,
            project=project or self.project or None,
            domain=domain or self.domain or None,
            time_partition=tp,
            partitions=p_obj or self.partitions,
        )
        return aq

    @property
    def concrete_artifact_id(self) -> art_id.ArtifactID:
        # This property is used when you want to ensure that this is a materialized artifact, all fields are known.
        if self.name is None or self.project is None or self.domain is None or self.version is None:
            raise ValueError("Cannot create artifact id without name, project, domain, version")
        return self.to_id_idl()

    def embed_as_query(
        self,
        bindings: typing.List[Artifact],
        partition: Optional[str] = None,
        bind_to_time_partition: Optional[bool] = None,
        expr: Optional[str] = None,
    ) -> art_id.ArtifactQuery:
        """
        This should only be called in the context of a Trigger
        :param bindings: The list of artifacts in trigger_on
        :param partition: Can embed a time partition
        :param bind_to_time_partition: Set to true if you want to bind to a time partition
        :param expr: Only valid if there's a time partition.
        """
        # Find self in the list, raises ValueError if not there.
        idx = bindings.index(self)
        aq = art_id.ArtifactQuery(
            binding=art_id.ArtifactBindingData(
                index=idx,
                partition_key=partition,
                bind_to_time_partition=bind_to_time_partition,
                transform=str(expr) if expr and (partition or bind_to_time_partition) else None,
            )
        )

        return aq

    def to_id_idl(self) -> art_id.ArtifactID:
        """
        Converts this object to the IDL representation.
        This is here instead of translator because it's in the interface, a relatively simple proto object
        that's exposed to the user.
        """
        p = Serializer.partitions_to_idl(self.partitions)
        tp = Serializer.time_partition_to_idl(self.time_partition) if self.time_partitioned else None

        i = art_id.ArtifactID(
            artifact_key=art_id.ArtifactKey(
                project=self.project,
                domain=self.domain,
                name=self.name,
            ),
            version=self.version,
            partitions=p,
            time_partition=tp,
        )

        return i


class ArtifactSerializationHandler(typing.Protocol):
    """
    This protocol defines the interface for serializing artifact-related entities down to Flyte IDL.
    """

    def partitions_to_idl(self, p: Optional[Partitions], **kwargs) -> Optional[art_id.Partitions]:
        ...

    def time_partition_to_idl(self, tp: Optional[TimePartition], **kwargs) -> Optional[art_id.TimePartition]:
        ...

    def artifact_query_to_idl(self, aq: ArtifactQuery, **kwargs) -> art_id.ArtifactQuery:
        ...


class DefaultArtifactSerializationHandler(ArtifactSerializationHandler):
    def partitions_to_idl(self, p: Optional[Partitions], **kwargs) -> Optional[art_id.Partitions]:
        if p and p.partitions:
            pp = {}
            for k, v in p.partitions.items():
                if v.value is None:
                    # For specifying partitions in the Variable partial id
                    pp[k] = art_id.LabelValue(static_value="")
                else:
                    pp[k] = v.value
            return art_id.Partitions(value=pp)
        return None

    def time_partition_to_idl(self, tp: Optional[TimePartition], **kwargs) -> Optional[art_id.TimePartition]:
        if tp:
            return art_id.TimePartition(value=tp.value)
        return None

    def artifact_query_to_idl(self, aq: ArtifactQuery, **kwargs) -> art_id.ArtifactQuery:
        ak = art_id.ArtifactKey(
            name=aq.name,
            project=aq.project,
            domain=aq.domain,
        )

        p = self.partitions_to_idl(aq.partitions)
        tp = self.time_partition_to_idl(aq.time_partition)

        i = art_id.ArtifactID(
            artifact_key=ak,
            partitions=p,
            time_partition=tp,
        )

        aq = art_id.ArtifactQuery(
            artifact_id=i,
        )

        return aq


class Serializer(object):
    serializer: ArtifactSerializationHandler = DefaultArtifactSerializationHandler()

    @classmethod
    def register_serializer(cls, serializer: ArtifactSerializationHandler):
        cls.serializer = serializer

    @classmethod
    def partitions_to_idl(cls, p: Optional[Partitions], **kwargs) -> Optional[art_id.Partitions]:
        return cls.serializer.partitions_to_idl(p, **kwargs)

    @classmethod
    def time_partition_to_idl(cls, tp: TimePartition, **kwargs) -> Optional[art_id.TimePartition]:
        return cls.serializer.time_partition_to_idl(tp, **kwargs)

    @classmethod
    def artifact_query_to_idl(cls, aq: ArtifactQuery, **kwargs) -> art_id.ArtifactQuery:
        return cls.serializer.artifact_query_to_idl(aq, **kwargs)

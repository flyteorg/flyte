from typing import Dict, List, Optional, Union

from flyteidl.plugins.sagemaker import parameter_ranges_pb2 as _idl_parameter_ranges

from flytekit.exceptions import user
from flytekit.models import common as _common


class HyperparameterScalingType(object):
    AUTO = _idl_parameter_ranges.HyperparameterScalingType.AUTO
    LINEAR = _idl_parameter_ranges.HyperparameterScalingType.LINEAR
    LOGARITHMIC = _idl_parameter_ranges.HyperparameterScalingType.LOGARITHMIC
    REVERSELOGARITHMIC = _idl_parameter_ranges.HyperparameterScalingType.REVERSELOGARITHMIC


class ContinuousParameterRange(_common.FlyteIdlEntity):
    def __init__(
        self,
        max_value: float,
        min_value: float,
        scaling_type: int,
    ):
        """

        :param float max_value:
        :param float min_value:
        :param int scaling_type:
        """
        self._max_value = max_value
        self._min_value = min_value
        self._scaling_type = scaling_type

    @property
    def max_value(self) -> float:
        """

        :rtype: float
        """
        return self._max_value

    @property
    def min_value(self) -> float:
        """

        :rtype: float
        """
        return self._min_value

    @property
    def scaling_type(self) -> int:
        """
        enum value from HyperparameterScalingType
        :rtype: int
        """
        return self._scaling_type

    def to_flyte_idl(self) -> _idl_parameter_ranges.ContinuousParameterRange:
        """
        :rtype: _idl_parameter_ranges.ContinuousParameterRange
        """

        return _idl_parameter_ranges.ContinuousParameterRange(
            max_value=self._max_value,
            min_value=self._min_value,
            scaling_type=self.scaling_type,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _idl_parameter_ranges.ContinuousParameterRange):
        """

        :param pb2_object:
        :rtype: ContinuousParameterRange
        """
        return cls(
            max_value=pb2_object.max_value,
            min_value=pb2_object.min_value,
            scaling_type=pb2_object.scaling_type,
        )


class IntegerParameterRange(_common.FlyteIdlEntity):
    def __init__(
        self,
        max_value: int,
        min_value: int,
        scaling_type: int,
    ):
        """
        :param int max_value:
        :param int min_value:
        :param int scaling_type:
        """
        self._max_value = max_value
        self._min_value = min_value
        self._scaling_type = scaling_type

    @property
    def max_value(self) -> int:
        """
        :rtype: int
        """
        return self._max_value

    @property
    def min_value(self) -> int:
        """

        :rtype: int
        """
        return self._min_value

    @property
    def scaling_type(self) -> int:
        """
        enum value from HyperparameterScalingType
        :rtype: int
        """
        return self._scaling_type

    def to_flyte_idl(self) -> _idl_parameter_ranges.IntegerParameterRange:
        """
        :rtype: _idl_parameter_ranges.IntegerParameterRange
        """
        return _idl_parameter_ranges.IntegerParameterRange(
            max_value=self._max_value,
            min_value=self._min_value,
            scaling_type=self.scaling_type,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _idl_parameter_ranges.IntegerParameterRange):
        """

        :param pb2_object:
        :rtype: IntegerParameterRange
        """
        return cls(
            max_value=pb2_object.max_value,
            min_value=pb2_object.min_value,
            scaling_type=pb2_object.scaling_type,
        )


class CategoricalParameterRange(_common.FlyteIdlEntity):
    def __init__(
        self,
        values: List[str],
    ):
        """

        :param List[str] values: list of strings representing categorical values
        """
        self._values = values

    @property
    def values(self) -> List[str]:
        """
        :rtype: List[str]
        """
        return self._values

    def to_flyte_idl(self) -> _idl_parameter_ranges.CategoricalParameterRange:
        """
        :rtype: _idl_parameter_ranges.CategoricalParameterRange
        """
        return _idl_parameter_ranges.CategoricalParameterRange(values=self._values)

    @classmethod
    def from_flyte_idl(cls, pb2_object: _idl_parameter_ranges.CategoricalParameterRange):
        """

        :param pb2_object:
        :rtype: CategoricalParameterRange
        """
        return cls(values=[v for v in pb2_object.values])


class ParameterRanges(_common.FlyteIdlEntity):
    def __init__(
        self,
        parameter_range_map: Dict[str, _common.FlyteIdlEntity],
    ):
        self._parameter_range_map = parameter_range_map

    def to_flyte_idl(self) -> _idl_parameter_ranges.ParameterRanges:
        """

        :rtype: _idl_parameter_ranges.ParameterRanges
        """
        converted = {}
        for k, v in self._parameter_range_map.items():
            if isinstance(v, IntegerParameterRange):
                converted[k] = _idl_parameter_ranges.ParameterRangeOneOf(integer_parameter_range=v.to_flyte_idl())
            elif isinstance(v, ContinuousParameterRange):
                converted[k] = _idl_parameter_ranges.ParameterRangeOneOf(continuous_parameter_range=v.to_flyte_idl())
            elif isinstance(v, CategoricalParameterRange):
                converted[k] = _idl_parameter_ranges.ParameterRangeOneOf(categorical_parameter_range=v.to_flyte_idl())
            else:
                raise user.FlyteTypeException(
                    received_type=type(v),
                    expected_type=type(
                        Union[IntegerParameterRange, ContinuousParameterRange, CategoricalParameterRange]
                    ),
                )

        return _idl_parameter_ranges.ParameterRanges(
            parameter_range_map=converted,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _idl_parameter_ranges.ParameterRanges):
        """

        :param pb2_object:
        :rtype: ParameterRanges
        """
        converted = {}
        for k, v in pb2_object.parameter_range_map.items():
            if v.HasField("continuous_parameter_range"):
                converted[k] = ContinuousParameterRange.from_flyte_idl(v.continuous_parameter_range)
            elif v.HasField("integer_parameter_range"):
                converted[k] = IntegerParameterRange.from_flyte_idl(v.integer_parameter_range)
            else:
                converted[k] = CategoricalParameterRange.from_flyte_idl(v.categorical_parameter_range)

        return cls(
            parameter_range_map=converted,
        )


class ParameterRangeOneOf(_common.FlyteIdlEntity):
    def __init__(self, param: Union[IntegerParameterRange, ContinuousParameterRange, CategoricalParameterRange]):
        """
        Initializes a new ParameterRangeOneOf.

        :param Union[IntegerParameterRange, ContinuousParameterRange, CategoricalParameterRange] param: One of the
            supported parameter ranges.
        """
        self._integer_parameter_range = param if isinstance(param, IntegerParameterRange) else None
        self._continuous_parameter_range = param if isinstance(param, ContinuousParameterRange) else None
        self._categorical_parameter_range = param if isinstance(param, CategoricalParameterRange) else None

    @property
    def integer_parameter_range(self) -> Optional[IntegerParameterRange]:
        """
        Retrieves the integer parameter range if one is set. None otherwise.
        :rtype: Optional[IntegerParameterRange]
        """
        if self._integer_parameter_range:
            return self._integer_parameter_range

        return None

    @property
    def continuous_parameter_range(self) -> Optional[ContinuousParameterRange]:
        """
        Retrieves the continuous parameter range if one is set. None otherwise.
        :rtype: Optional[ContinuousParameterRange]
        """
        if self._continuous_parameter_range:
            return self._continuous_parameter_range

        return None

    @property
    def categorical_parameter_range(self) -> Optional[CategoricalParameterRange]:
        """
        Retrieves the categorical parameter range if one is set. None otherwise.
        :rtype: Optional[CategoricalParameterRange]
        """
        if self._categorical_parameter_range:
            return self._categorical_parameter_range

        return None

    def to_flyte_idl(self) -> _idl_parameter_ranges.ParameterRangeOneOf:
        return _idl_parameter_ranges.ParameterRangeOneOf(
            integer_parameter_range=self.integer_parameter_range.to_flyte_idl()
            if self.integer_parameter_range
            else None,
            continuous_parameter_range=self.continuous_parameter_range.to_flyte_idl()
            if self.continuous_parameter_range
            else None,
            categorical_parameter_range=self.categorical_parameter_range.to_flyte_idl()
            if self.categorical_parameter_range
            else None,
        )

    @classmethod
    def from_flyte_idl(
        cls,
        pb_object: Union[
            _idl_parameter_ranges.ParameterRangeOneOf,
            _idl_parameter_ranges.IntegerParameterRange,
            _idl_parameter_ranges.ContinuousParameterRange,
            _idl_parameter_ranges.CategoricalParameterRange,
        ],
    ):
        param = None
        if isinstance(pb_object, _idl_parameter_ranges.ParameterRangeOneOf):
            if pb_object.HasField("continuous_parameter_range"):
                param = ContinuousParameterRange.from_flyte_idl(pb_object.continuous_parameter_range)
            elif pb_object.HasField("integer_parameter_range"):
                param = IntegerParameterRange.from_flyte_idl(pb_object.integer_parameter_range)
            elif pb_object.HasField("categorical_parameter_range"):
                param = CategoricalParameterRange.from_flyte_idl(pb_object.categorical_parameter_range)
        elif isinstance(pb_object, _idl_parameter_ranges.IntegerParameterRange):
            param = IntegerParameterRange.from_flyte_idl(pb_object)
        elif isinstance(pb_object, _idl_parameter_ranges.ContinuousParameterRange):
            param = ContinuousParameterRange.from_flyte_idl(pb_object)
        elif isinstance(pb_object, _idl_parameter_ranges.CategoricalParameterRange):
            param = CategoricalParameterRange.from_flyte_idl(pb_object)

        return cls(param=param)

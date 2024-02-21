from flytekit.models.common import FlyteIdlEntity as _FlyteIdlEntity


class FilterList(_FlyteIdlEntity):
    def __init__(self, filter_list):
        """
        :param list[Filter] filter_list: List of filters to AND together
        """
        self._filter_list = filter_list

    def to_flyte_idl(self):
        """
        For supporting the auto-generated REST API, filters must be dumped to a string for representation as GET params.
        :rtype: Text
        """
        return "+".join([f.to_flyte_idl() for f in self._filter_list])

    @classmethod
    def from_flyte_idl(cls):
        raise NotImplementedError("Filters are never recovered from a protobuf.")


class Filter(_FlyteIdlEntity):
    _comparator = "nil"

    def __init__(self, key, value):
        """
        :param Text key: The name of the field to compare against
        :param Text value: The textual value of the field to compare against
        """
        self._key = key
        self._value = value

    def to_flyte_idl(self):
        """
        For supporting the auto-generated REST API, filters must be dumped to a string for representation as GET params.
        :rtype: Text
        """
        return "{}({},{})".format(type(self)._comparator, self._key, self._value)

    @classmethod
    def from_flyte_idl(cls):
        raise NotImplementedError("Filters are never recovered from a protobuf.")

    @classmethod
    def from_python_std(cls, string):
        """
        :param Text string:
        :rtype: Filter
        """
        if string.startswith("eq("):
            return Equal._parse_from_string(string)
        elif string.startswith("ne("):
            return NotEqual._parse_from_string(string)
        elif string.startswith("gt("):
            return GreaterThan._parse_from_string(string)
        elif string.startswith("gte("):
            return GreaterThanOrEqual._parse_from_string(string)
        elif string.startswith("lt("):
            return LessThan._parse_from_string(string)
        elif string.startswith("lte("):
            return LessThanOrEqual._parse_from_string(string)
        elif string.startswith("contains("):
            return Contains._parse_from_string(string)
        elif string.startswith("value_in("):
            return ValueIn._parse_from_string(string)
        else:
            raise ValueError("'{}' could not be parsed into a filter.".format(string))

    @classmethod
    def _parse_from_string(cls, string):
        """
        :param Text string:
        :rtype: Filter
        """
        stripped = string[len(cls._comparator) + 1 :]
        if stripped[-1] != ")":
            raise ValueError("Filter could not be parsed because {} did not end with a ')'".format(string))
        split = stripped[:-1].split(",")
        if len(split) != 2:
            raise ValueError("Filter must be expressed as a key, value tuple like 'eq(abc,def)'")
        key = split[0].strip()
        value = split[1].strip()
        return cls(key, cls._parse_value(value))

    @classmethod
    def _parse_value(cls, value):
        return value


class Equal(Filter):
    _comparator = "eq"


class NotEqual(Filter):
    _comparator = "ne"


class GreaterThan(Filter):
    _comparator = "gt"


class GreaterThanOrEqual(Filter):
    _comparator = "gte"


class LessThan(Filter):
    _comparator = "lt"


class LessThanOrEqual(Filter):
    _comparator = "lte"


class SetFilter(Filter):
    def __init__(self, key, values):
        """
        :param Text key:  The name of the field to compare against
        :param list[Text] values:  A list of textual values to compare.
        """
        super(SetFilter, self).__init__(key, ";".join(values))

    @classmethod
    def _parse_value(cls, value):
        return value.split(";")


class Contains(SetFilter):
    _comparator = "contains"


class ValueIn(SetFilter):
    _comparator = "value_in"

# coding: utf-8

"""
    flyteidl/service/admin.proto

    No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)  # noqa: E501

    OpenAPI spec version: version not set
    
    Generated by: https://github.com/swagger-api/swagger-codegen.git
"""


import pprint
import re  # noqa: F401

import six

from flyteadmin.configuration import Configuration


class CoreBindingData(object):
    """NOTE: This class is auto generated by the swagger code generator program.

    Do not edit the class manually.
    """

    """
    Attributes:
      swagger_types (dict): The key is attribute name
                            and the value is attribute type.
      attribute_map (dict): The key is attribute name
                            and the value is json key in definition.
    """
    swagger_types = {
        'scalar': 'CoreScalar',
        'collection': 'CoreBindingDataCollection',
        'promise': 'CoreOutputReference',
        'map': 'CoreBindingDataMap',
        'union': 'CoreUnionInfo'
    }

    attribute_map = {
        'scalar': 'scalar',
        'collection': 'collection',
        'promise': 'promise',
        'map': 'map',
        'union': 'union'
    }

    def __init__(self, scalar=None, collection=None, promise=None, map=None, union=None, _configuration=None):  # noqa: E501
        """CoreBindingData - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._scalar = None
        self._collection = None
        self._promise = None
        self._map = None
        self._union = None
        self.discriminator = None

        if scalar is not None:
            self.scalar = scalar
        if collection is not None:
            self.collection = collection
        if promise is not None:
            self.promise = promise
        if map is not None:
            self.map = map
        if union is not None:
            self.union = union

    @property
    def scalar(self):
        """Gets the scalar of this CoreBindingData.  # noqa: E501

        A simple scalar value.  # noqa: E501

        :return: The scalar of this CoreBindingData.  # noqa: E501
        :rtype: CoreScalar
        """
        return self._scalar

    @scalar.setter
    def scalar(self, scalar):
        """Sets the scalar of this CoreBindingData.

        A simple scalar value.  # noqa: E501

        :param scalar: The scalar of this CoreBindingData.  # noqa: E501
        :type: CoreScalar
        """

        self._scalar = scalar

    @property
    def collection(self):
        """Gets the collection of this CoreBindingData.  # noqa: E501

        A collection of binding data. This allows nesting of binding data to any number of levels.  # noqa: E501

        :return: The collection of this CoreBindingData.  # noqa: E501
        :rtype: CoreBindingDataCollection
        """
        return self._collection

    @collection.setter
    def collection(self, collection):
        """Sets the collection of this CoreBindingData.

        A collection of binding data. This allows nesting of binding data to any number of levels.  # noqa: E501

        :param collection: The collection of this CoreBindingData.  # noqa: E501
        :type: CoreBindingDataCollection
        """

        self._collection = collection

    @property
    def promise(self):
        """Gets the promise of this CoreBindingData.  # noqa: E501

        References an output promised by another node.  # noqa: E501

        :return: The promise of this CoreBindingData.  # noqa: E501
        :rtype: CoreOutputReference
        """
        return self._promise

    @promise.setter
    def promise(self, promise):
        """Sets the promise of this CoreBindingData.

        References an output promised by another node.  # noqa: E501

        :param promise: The promise of this CoreBindingData.  # noqa: E501
        :type: CoreOutputReference
        """

        self._promise = promise

    @property
    def map(self):
        """Gets the map of this CoreBindingData.  # noqa: E501

        A map of bindings. The key is always a string.  # noqa: E501

        :return: The map of this CoreBindingData.  # noqa: E501
        :rtype: CoreBindingDataMap
        """
        return self._map

    @map.setter
    def map(self, map):
        """Sets the map of this CoreBindingData.

        A map of bindings. The key is always a string.  # noqa: E501

        :param map: The map of this CoreBindingData.  # noqa: E501
        :type: CoreBindingDataMap
        """

        self._map = map

    @property
    def union(self):
        """Gets the union of this CoreBindingData.  # noqa: E501


        :return: The union of this CoreBindingData.  # noqa: E501
        :rtype: CoreUnionInfo
        """
        return self._union

    @union.setter
    def union(self, union):
        """Sets the union of this CoreBindingData.


        :param union: The union of this CoreBindingData.  # noqa: E501
        :type: CoreUnionInfo
        """

        self._union = union

    def to_dict(self):
        """Returns the model properties as a dict"""
        result = {}

        for attr, _ in six.iteritems(self.swagger_types):
            value = getattr(self, attr)
            if isinstance(value, list):
                result[attr] = list(map(
                    lambda x: x.to_dict() if hasattr(x, "to_dict") else x,
                    value
                ))
            elif hasattr(value, "to_dict"):
                result[attr] = value.to_dict()
            elif isinstance(value, dict):
                result[attr] = dict(map(
                    lambda item: (item[0], item[1].to_dict())
                    if hasattr(item[1], "to_dict") else item,
                    value.items()
                ))
            else:
                result[attr] = value
        if issubclass(CoreBindingData, dict):
            for key, value in self.items():
                result[key] = value

        return result

    def to_str(self):
        """Returns the string representation of the model"""
        return pprint.pformat(self.to_dict())

    def __repr__(self):
        """For `print` and `pprint`"""
        return self.to_str()

    def __eq__(self, other):
        """Returns true if both objects are equal"""
        if not isinstance(other, CoreBindingData):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, CoreBindingData):
            return True

        return self.to_dict() != other.to_dict()

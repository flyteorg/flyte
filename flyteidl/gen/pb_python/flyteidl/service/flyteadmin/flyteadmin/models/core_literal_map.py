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


class CoreLiteralMap(object):
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
        'literals': 'dict(str, CoreLiteral)'
    }

    attribute_map = {
        'literals': 'literals'
    }

    def __init__(self, literals=None, _configuration=None):  # noqa: E501
        """CoreLiteralMap - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._literals = None
        self.discriminator = None

        if literals is not None:
            self.literals = literals

    @property
    def literals(self):
        """Gets the literals of this CoreLiteralMap.  # noqa: E501


        :return: The literals of this CoreLiteralMap.  # noqa: E501
        :rtype: dict(str, CoreLiteral)
        """
        return self._literals

    @literals.setter
    def literals(self, literals):
        """Sets the literals of this CoreLiteralMap.


        :param literals: The literals of this CoreLiteralMap.  # noqa: E501
        :type: dict(str, CoreLiteral)
        """

        self._literals = literals

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
        if issubclass(CoreLiteralMap, dict):
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
        if not isinstance(other, CoreLiteralMap):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, CoreLiteralMap):
            return True

        return self.to_dict() != other.to_dict()

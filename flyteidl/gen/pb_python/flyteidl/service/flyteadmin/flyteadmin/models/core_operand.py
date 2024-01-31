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


class CoreOperand(object):
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
        'primitive': 'CorePrimitive',
        'var': 'str',
        'scalar': 'CoreScalar'
    }

    attribute_map = {
        'primitive': 'primitive',
        'var': 'var',
        'scalar': 'scalar'
    }

    def __init__(self, primitive=None, var=None, scalar=None, _configuration=None):  # noqa: E501
        """CoreOperand - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._primitive = None
        self._var = None
        self._scalar = None
        self.discriminator = None

        if primitive is not None:
            self.primitive = primitive
        if var is not None:
            self.var = var
        if scalar is not None:
            self.scalar = scalar

    @property
    def primitive(self):
        """Gets the primitive of this CoreOperand.  # noqa: E501


        :return: The primitive of this CoreOperand.  # noqa: E501
        :rtype: CorePrimitive
        """
        return self._primitive

    @primitive.setter
    def primitive(self, primitive):
        """Sets the primitive of this CoreOperand.


        :param primitive: The primitive of this CoreOperand.  # noqa: E501
        :type: CorePrimitive
        """

        self._primitive = primitive

    @property
    def var(self):
        """Gets the var of this CoreOperand.  # noqa: E501


        :return: The var of this CoreOperand.  # noqa: E501
        :rtype: str
        """
        return self._var

    @var.setter
    def var(self, var):
        """Sets the var of this CoreOperand.


        :param var: The var of this CoreOperand.  # noqa: E501
        :type: str
        """

        self._var = var

    @property
    def scalar(self):
        """Gets the scalar of this CoreOperand.  # noqa: E501


        :return: The scalar of this CoreOperand.  # noqa: E501
        :rtype: CoreScalar
        """
        return self._scalar

    @scalar.setter
    def scalar(self, scalar):
        """Sets the scalar of this CoreOperand.


        :param scalar: The scalar of this CoreOperand.  # noqa: E501
        :type: CoreScalar
        """

        self._scalar = scalar

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
        if issubclass(CoreOperand, dict):
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
        if not isinstance(other, CoreOperand):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, CoreOperand):
            return True

        return self.to_dict() != other.to_dict()

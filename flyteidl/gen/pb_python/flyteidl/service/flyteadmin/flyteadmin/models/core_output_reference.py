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


class CoreOutputReference(object):
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
        'node_id': 'str',
        'var': 'str',
        'attr_path': 'list[CorePromiseAttribute]'
    }

    attribute_map = {
        'node_id': 'node_id',
        'var': 'var',
        'attr_path': 'attr_path'
    }

    def __init__(self, node_id=None, var=None, attr_path=None, _configuration=None):  # noqa: E501
        """CoreOutputReference - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._node_id = None
        self._var = None
        self._attr_path = None
        self.discriminator = None

        if node_id is not None:
            self.node_id = node_id
        if var is not None:
            self.var = var
        if attr_path is not None:
            self.attr_path = attr_path

    @property
    def node_id(self):
        """Gets the node_id of this CoreOutputReference.  # noqa: E501

        Node id must exist at the graph layer.  # noqa: E501

        :return: The node_id of this CoreOutputReference.  # noqa: E501
        :rtype: str
        """
        return self._node_id

    @node_id.setter
    def node_id(self, node_id):
        """Sets the node_id of this CoreOutputReference.

        Node id must exist at the graph layer.  # noqa: E501

        :param node_id: The node_id of this CoreOutputReference.  # noqa: E501
        :type: str
        """

        self._node_id = node_id

    @property
    def var(self):
        """Gets the var of this CoreOutputReference.  # noqa: E501

        Variable name must refer to an output variable for the node.  # noqa: E501

        :return: The var of this CoreOutputReference.  # noqa: E501
        :rtype: str
        """
        return self._var

    @var.setter
    def var(self, var):
        """Sets the var of this CoreOutputReference.

        Variable name must refer to an output variable for the node.  # noqa: E501

        :param var: The var of this CoreOutputReference.  # noqa: E501
        :type: str
        """

        self._var = var

    @property
    def attr_path(self):
        """Gets the attr_path of this CoreOutputReference.  # noqa: E501


        :return: The attr_path of this CoreOutputReference.  # noqa: E501
        :rtype: list[CorePromiseAttribute]
        """
        return self._attr_path

    @attr_path.setter
    def attr_path(self, attr_path):
        """Sets the attr_path of this CoreOutputReference.


        :param attr_path: The attr_path of this CoreOutputReference.  # noqa: E501
        :type: list[CorePromiseAttribute]
        """

        self._attr_path = attr_path

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
        if issubclass(CoreOutputReference, dict):
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
        if not isinstance(other, CoreOutputReference):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, CoreOutputReference):
            return True

        return self.to_dict() != other.to_dict()

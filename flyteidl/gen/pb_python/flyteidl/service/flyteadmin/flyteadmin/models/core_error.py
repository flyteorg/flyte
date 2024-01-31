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


class CoreError(object):
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
        'failed_node_id': 'str',
        'message': 'str'
    }

    attribute_map = {
        'failed_node_id': 'failed_node_id',
        'message': 'message'
    }

    def __init__(self, failed_node_id=None, message=None, _configuration=None):  # noqa: E501
        """CoreError - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._failed_node_id = None
        self._message = None
        self.discriminator = None

        if failed_node_id is not None:
            self.failed_node_id = failed_node_id
        if message is not None:
            self.message = message

    @property
    def failed_node_id(self):
        """Gets the failed_node_id of this CoreError.  # noqa: E501

        The node id that threw the error.  # noqa: E501

        :return: The failed_node_id of this CoreError.  # noqa: E501
        :rtype: str
        """
        return self._failed_node_id

    @failed_node_id.setter
    def failed_node_id(self, failed_node_id):
        """Sets the failed_node_id of this CoreError.

        The node id that threw the error.  # noqa: E501

        :param failed_node_id: The failed_node_id of this CoreError.  # noqa: E501
        :type: str
        """

        self._failed_node_id = failed_node_id

    @property
    def message(self):
        """Gets the message of this CoreError.  # noqa: E501

        Error message thrown.  # noqa: E501

        :return: The message of this CoreError.  # noqa: E501
        :rtype: str
        """
        return self._message

    @message.setter
    def message(self, message):
        """Sets the message of this CoreError.

        Error message thrown.  # noqa: E501

        :param message: The message of this CoreError.  # noqa: E501
        :type: str
        """

        self._message = message

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
        if issubclass(CoreError, dict):
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
        if not isinstance(other, CoreError):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, CoreError):
            return True

        return self.to_dict() != other.to_dict()

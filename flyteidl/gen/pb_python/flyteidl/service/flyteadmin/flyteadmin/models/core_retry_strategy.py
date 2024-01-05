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


class CoreRetryStrategy(object):
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
        'retries': 'int'
    }

    attribute_map = {
        'retries': 'retries'
    }

    def __init__(self, retries=None):  # noqa: E501
        """CoreRetryStrategy - a model defined in Swagger"""  # noqa: E501

        self._retries = None
        self.discriminator = None

        if retries is not None:
            self.retries = retries

    @property
    def retries(self):
        """Gets the retries of this CoreRetryStrategy.  # noqa: E501

        Number of retries. Retries will be consumed when the job fails with a recoverable error. The number of retries must be less than or equals to 10.  # noqa: E501

        :return: The retries of this CoreRetryStrategy.  # noqa: E501
        :rtype: int
        """
        return self._retries

    @retries.setter
    def retries(self, retries):
        """Sets the retries of this CoreRetryStrategy.

        Number of retries. Retries will be consumed when the job fails with a recoverable error. The number of retries must be less than or equals to 10.  # noqa: E501

        :param retries: The retries of this CoreRetryStrategy.  # noqa: E501
        :type: int
        """

        self._retries = retries

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
        if issubclass(CoreRetryStrategy, dict):
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
        if not isinstance(other, CoreRetryStrategy):
            return False

        return self.__dict__ == other.__dict__

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        return not self == other

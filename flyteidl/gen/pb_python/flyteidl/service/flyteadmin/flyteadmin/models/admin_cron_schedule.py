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


class AdminCronSchedule(object):
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
        'schedule': 'str',
        'offset': 'str'
    }

    attribute_map = {
        'schedule': 'schedule',
        'offset': 'offset'
    }

    def __init__(self, schedule=None, offset=None, _configuration=None):  # noqa: E501
        """AdminCronSchedule - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._schedule = None
        self._offset = None
        self.discriminator = None

        if schedule is not None:
            self.schedule = schedule
        if offset is not None:
            self.offset = offset

    @property
    def schedule(self):
        """Gets the schedule of this AdminCronSchedule.  # noqa: E501


        :return: The schedule of this AdminCronSchedule.  # noqa: E501
        :rtype: str
        """
        return self._schedule

    @schedule.setter
    def schedule(self, schedule):
        """Sets the schedule of this AdminCronSchedule.


        :param schedule: The schedule of this AdminCronSchedule.  # noqa: E501
        :type: str
        """

        self._schedule = schedule

    @property
    def offset(self):
        """Gets the offset of this AdminCronSchedule.  # noqa: E501


        :return: The offset of this AdminCronSchedule.  # noqa: E501
        :rtype: str
        """
        return self._offset

    @offset.setter
    def offset(self, offset):
        """Sets the offset of this AdminCronSchedule.


        :param offset: The offset of this AdminCronSchedule.  # noqa: E501
        :type: str
        """

        self._offset = offset

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
        if issubclass(AdminCronSchedule, dict):
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
        if not isinstance(other, AdminCronSchedule):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, AdminCronSchedule):
            return True

        return self.to_dict() != other.to_dict()

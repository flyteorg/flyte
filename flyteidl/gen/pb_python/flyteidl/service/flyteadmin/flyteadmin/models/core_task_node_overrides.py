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


class CoreTaskNodeOverrides(object):
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
        'resources': 'CoreResources',
        'extended_resources': 'CoreExtendedResources'
    }

    attribute_map = {
        'resources': 'resources',
        'extended_resources': 'extended_resources'
    }

    def __init__(self, resources=None, extended_resources=None, _configuration=None):  # noqa: E501
        """CoreTaskNodeOverrides - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._resources = None
        self._extended_resources = None
        self.discriminator = None

        if resources is not None:
            self.resources = resources
        if extended_resources is not None:
            self.extended_resources = extended_resources

    @property
    def resources(self):
        """Gets the resources of this CoreTaskNodeOverrides.  # noqa: E501

        A customizable interface to convey resources requested for a task container.  # noqa: E501

        :return: The resources of this CoreTaskNodeOverrides.  # noqa: E501
        :rtype: CoreResources
        """
        return self._resources

    @resources.setter
    def resources(self, resources):
        """Sets the resources of this CoreTaskNodeOverrides.

        A customizable interface to convey resources requested for a task container.  # noqa: E501

        :param resources: The resources of this CoreTaskNodeOverrides.  # noqa: E501
        :type: CoreResources
        """

        self._resources = resources

    @property
    def extended_resources(self):
        """Gets the extended_resources of this CoreTaskNodeOverrides.  # noqa: E501

        Overrides for all non-standard resources, not captured by v1.ResourceRequirements, to allocate to a task.  # noqa: E501

        :return: The extended_resources of this CoreTaskNodeOverrides.  # noqa: E501
        :rtype: CoreExtendedResources
        """
        return self._extended_resources

    @extended_resources.setter
    def extended_resources(self, extended_resources):
        """Sets the extended_resources of this CoreTaskNodeOverrides.

        Overrides for all non-standard resources, not captured by v1.ResourceRequirements, to allocate to a task.  # noqa: E501

        :param extended_resources: The extended_resources of this CoreTaskNodeOverrides.  # noqa: E501
        :type: CoreExtendedResources
        """

        self._extended_resources = extended_resources

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
        if issubclass(CoreTaskNodeOverrides, dict):
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
        if not isinstance(other, CoreTaskNodeOverrides):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, CoreTaskNodeOverrides):
            return True

        return self.to_dict() != other.to_dict()

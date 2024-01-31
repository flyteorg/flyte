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


class AdminPluginOverride(object):
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
        'task_type': 'str',
        'plugin_id': 'list[str]',
        'missing_plugin_behavior': 'PluginOverrideMissingPluginBehavior'
    }

    attribute_map = {
        'task_type': 'task_type',
        'plugin_id': 'plugin_id',
        'missing_plugin_behavior': 'missing_plugin_behavior'
    }

    def __init__(self, task_type=None, plugin_id=None, missing_plugin_behavior=None, _configuration=None):  # noqa: E501
        """AdminPluginOverride - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._task_type = None
        self._plugin_id = None
        self._missing_plugin_behavior = None
        self.discriminator = None

        if task_type is not None:
            self.task_type = task_type
        if plugin_id is not None:
            self.plugin_id = plugin_id
        if missing_plugin_behavior is not None:
            self.missing_plugin_behavior = missing_plugin_behavior

    @property
    def task_type(self):
        """Gets the task_type of this AdminPluginOverride.  # noqa: E501

        A predefined yet extensible Task type identifier.  # noqa: E501

        :return: The task_type of this AdminPluginOverride.  # noqa: E501
        :rtype: str
        """
        return self._task_type

    @task_type.setter
    def task_type(self, task_type):
        """Sets the task_type of this AdminPluginOverride.

        A predefined yet extensible Task type identifier.  # noqa: E501

        :param task_type: The task_type of this AdminPluginOverride.  # noqa: E501
        :type: str
        """

        self._task_type = task_type

    @property
    def plugin_id(self):
        """Gets the plugin_id of this AdminPluginOverride.  # noqa: E501

        A set of plugin ids which should handle tasks of this type instead of the default registered plugin. The list will be tried in order until a plugin is found with that id.  # noqa: E501

        :return: The plugin_id of this AdminPluginOverride.  # noqa: E501
        :rtype: list[str]
        """
        return self._plugin_id

    @plugin_id.setter
    def plugin_id(self, plugin_id):
        """Sets the plugin_id of this AdminPluginOverride.

        A set of plugin ids which should handle tasks of this type instead of the default registered plugin. The list will be tried in order until a plugin is found with that id.  # noqa: E501

        :param plugin_id: The plugin_id of this AdminPluginOverride.  # noqa: E501
        :type: list[str]
        """

        self._plugin_id = plugin_id

    @property
    def missing_plugin_behavior(self):
        """Gets the missing_plugin_behavior of this AdminPluginOverride.  # noqa: E501

        Defines the behavior when no plugin from the plugin_id list is not found.  # noqa: E501

        :return: The missing_plugin_behavior of this AdminPluginOverride.  # noqa: E501
        :rtype: PluginOverrideMissingPluginBehavior
        """
        return self._missing_plugin_behavior

    @missing_plugin_behavior.setter
    def missing_plugin_behavior(self, missing_plugin_behavior):
        """Sets the missing_plugin_behavior of this AdminPluginOverride.

        Defines the behavior when no plugin from the plugin_id list is not found.  # noqa: E501

        :param missing_plugin_behavior: The missing_plugin_behavior of this AdminPluginOverride.  # noqa: E501
        :type: PluginOverrideMissingPluginBehavior
        """

        self._missing_plugin_behavior = missing_plugin_behavior

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
        if issubclass(AdminPluginOverride, dict):
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
        if not isinstance(other, AdminPluginOverride):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, AdminPluginOverride):
            return True

        return self.to_dict() != other.to_dict()

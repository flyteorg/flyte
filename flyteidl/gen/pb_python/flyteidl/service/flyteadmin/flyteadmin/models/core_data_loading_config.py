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


class CoreDataLoadingConfig(object):
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
        'enabled': 'bool',
        'input_path': 'str',
        'output_path': 'str',
        'format': 'DataLoadingConfigLiteralMapFormat',
        'io_strategy': 'CoreIOStrategy'
    }

    attribute_map = {
        'enabled': 'enabled',
        'input_path': 'input_path',
        'output_path': 'output_path',
        'format': 'format',
        'io_strategy': 'io_strategy'
    }

    def __init__(self, enabled=None, input_path=None, output_path=None, format=None, io_strategy=None, _configuration=None):  # noqa: E501
        """CoreDataLoadingConfig - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._enabled = None
        self._input_path = None
        self._output_path = None
        self._format = None
        self._io_strategy = None
        self.discriminator = None

        if enabled is not None:
            self.enabled = enabled
        if input_path is not None:
            self.input_path = input_path
        if output_path is not None:
            self.output_path = output_path
        if format is not None:
            self.format = format
        if io_strategy is not None:
            self.io_strategy = io_strategy

    @property
    def enabled(self):
        """Gets the enabled of this CoreDataLoadingConfig.  # noqa: E501


        :return: The enabled of this CoreDataLoadingConfig.  # noqa: E501
        :rtype: bool
        """
        return self._enabled

    @enabled.setter
    def enabled(self, enabled):
        """Sets the enabled of this CoreDataLoadingConfig.


        :param enabled: The enabled of this CoreDataLoadingConfig.  # noqa: E501
        :type: bool
        """

        self._enabled = enabled

    @property
    def input_path(self):
        """Gets the input_path of this CoreDataLoadingConfig.  # noqa: E501


        :return: The input_path of this CoreDataLoadingConfig.  # noqa: E501
        :rtype: str
        """
        return self._input_path

    @input_path.setter
    def input_path(self, input_path):
        """Sets the input_path of this CoreDataLoadingConfig.


        :param input_path: The input_path of this CoreDataLoadingConfig.  # noqa: E501
        :type: str
        """

        self._input_path = input_path

    @property
    def output_path(self):
        """Gets the output_path of this CoreDataLoadingConfig.  # noqa: E501


        :return: The output_path of this CoreDataLoadingConfig.  # noqa: E501
        :rtype: str
        """
        return self._output_path

    @output_path.setter
    def output_path(self, output_path):
        """Sets the output_path of this CoreDataLoadingConfig.


        :param output_path: The output_path of this CoreDataLoadingConfig.  # noqa: E501
        :type: str
        """

        self._output_path = output_path

    @property
    def format(self):
        """Gets the format of this CoreDataLoadingConfig.  # noqa: E501


        :return: The format of this CoreDataLoadingConfig.  # noqa: E501
        :rtype: DataLoadingConfigLiteralMapFormat
        """
        return self._format

    @format.setter
    def format(self, format):
        """Sets the format of this CoreDataLoadingConfig.


        :param format: The format of this CoreDataLoadingConfig.  # noqa: E501
        :type: DataLoadingConfigLiteralMapFormat
        """

        self._format = format

    @property
    def io_strategy(self):
        """Gets the io_strategy of this CoreDataLoadingConfig.  # noqa: E501


        :return: The io_strategy of this CoreDataLoadingConfig.  # noqa: E501
        :rtype: CoreIOStrategy
        """
        return self._io_strategy

    @io_strategy.setter
    def io_strategy(self, io_strategy):
        """Sets the io_strategy of this CoreDataLoadingConfig.


        :param io_strategy: The io_strategy of this CoreDataLoadingConfig.  # noqa: E501
        :type: CoreIOStrategy
        """

        self._io_strategy = io_strategy

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
        if issubclass(CoreDataLoadingConfig, dict):
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
        if not isinstance(other, CoreDataLoadingConfig):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, CoreDataLoadingConfig):
            return True

        return self.to_dict() != other.to_dict()

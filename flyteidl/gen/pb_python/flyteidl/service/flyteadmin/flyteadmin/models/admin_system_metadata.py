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


class AdminSystemMetadata(object):
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
        'execution_cluster': 'str',
        'namespace': 'str'
    }

    attribute_map = {
        'execution_cluster': 'execution_cluster',
        'namespace': 'namespace'
    }

    def __init__(self, execution_cluster=None, namespace=None, _configuration=None):  # noqa: E501
        """AdminSystemMetadata - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._execution_cluster = None
        self._namespace = None
        self.discriminator = None

        if execution_cluster is not None:
            self.execution_cluster = execution_cluster
        if namespace is not None:
            self.namespace = namespace

    @property
    def execution_cluster(self):
        """Gets the execution_cluster of this AdminSystemMetadata.  # noqa: E501

        Which execution cluster this execution ran on.  # noqa: E501

        :return: The execution_cluster of this AdminSystemMetadata.  # noqa: E501
        :rtype: str
        """
        return self._execution_cluster

    @execution_cluster.setter
    def execution_cluster(self, execution_cluster):
        """Sets the execution_cluster of this AdminSystemMetadata.

        Which execution cluster this execution ran on.  # noqa: E501

        :param execution_cluster: The execution_cluster of this AdminSystemMetadata.  # noqa: E501
        :type: str
        """

        self._execution_cluster = execution_cluster

    @property
    def namespace(self):
        """Gets the namespace of this AdminSystemMetadata.  # noqa: E501

        Which kubernetes namespace the execution ran under.  # noqa: E501

        :return: The namespace of this AdminSystemMetadata.  # noqa: E501
        :rtype: str
        """
        return self._namespace

    @namespace.setter
    def namespace(self, namespace):
        """Sets the namespace of this AdminSystemMetadata.

        Which kubernetes namespace the execution ran under.  # noqa: E501

        :param namespace: The namespace of this AdminSystemMetadata.  # noqa: E501
        :type: str
        """

        self._namespace = namespace

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
        if issubclass(AdminSystemMetadata, dict):
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
        if not isinstance(other, AdminSystemMetadata):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, AdminSystemMetadata):
            return True

        return self.to_dict() != other.to_dict()

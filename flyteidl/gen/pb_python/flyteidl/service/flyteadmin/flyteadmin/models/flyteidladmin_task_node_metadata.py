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


class FlyteidladminTaskNodeMetadata(object):
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
        'cache_status': 'CoreCatalogCacheStatus',
        'catalog_key': 'CoreCatalogMetadata',
        'checkpoint_uri': 'str'
    }

    attribute_map = {
        'cache_status': 'cache_status',
        'catalog_key': 'catalog_key',
        'checkpoint_uri': 'checkpoint_uri'
    }

    def __init__(self, cache_status=None, catalog_key=None, checkpoint_uri=None, _configuration=None):  # noqa: E501
        """FlyteidladminTaskNodeMetadata - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._cache_status = None
        self._catalog_key = None
        self._checkpoint_uri = None
        self.discriminator = None

        if cache_status is not None:
            self.cache_status = cache_status
        if catalog_key is not None:
            self.catalog_key = catalog_key
        if checkpoint_uri is not None:
            self.checkpoint_uri = checkpoint_uri

    @property
    def cache_status(self):
        """Gets the cache_status of this FlyteidladminTaskNodeMetadata.  # noqa: E501

        Captures the status of caching for this execution.  # noqa: E501

        :return: The cache_status of this FlyteidladminTaskNodeMetadata.  # noqa: E501
        :rtype: CoreCatalogCacheStatus
        """
        return self._cache_status

    @cache_status.setter
    def cache_status(self, cache_status):
        """Sets the cache_status of this FlyteidladminTaskNodeMetadata.

        Captures the status of caching for this execution.  # noqa: E501

        :param cache_status: The cache_status of this FlyteidladminTaskNodeMetadata.  # noqa: E501
        :type: CoreCatalogCacheStatus
        """

        self._cache_status = cache_status

    @property
    def catalog_key(self):
        """Gets the catalog_key of this FlyteidladminTaskNodeMetadata.  # noqa: E501


        :return: The catalog_key of this FlyteidladminTaskNodeMetadata.  # noqa: E501
        :rtype: CoreCatalogMetadata
        """
        return self._catalog_key

    @catalog_key.setter
    def catalog_key(self, catalog_key):
        """Sets the catalog_key of this FlyteidladminTaskNodeMetadata.


        :param catalog_key: The catalog_key of this FlyteidladminTaskNodeMetadata.  # noqa: E501
        :type: CoreCatalogMetadata
        """

        self._catalog_key = catalog_key

    @property
    def checkpoint_uri(self):
        """Gets the checkpoint_uri of this FlyteidladminTaskNodeMetadata.  # noqa: E501


        :return: The checkpoint_uri of this FlyteidladminTaskNodeMetadata.  # noqa: E501
        :rtype: str
        """
        return self._checkpoint_uri

    @checkpoint_uri.setter
    def checkpoint_uri(self, checkpoint_uri):
        """Sets the checkpoint_uri of this FlyteidladminTaskNodeMetadata.


        :param checkpoint_uri: The checkpoint_uri of this FlyteidladminTaskNodeMetadata.  # noqa: E501
        :type: str
        """

        self._checkpoint_uri = checkpoint_uri

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
        if issubclass(FlyteidladminTaskNodeMetadata, dict):
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
        if not isinstance(other, FlyteidladminTaskNodeMetadata):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, FlyteidladminTaskNodeMetadata):
            return True

        return self.to_dict() != other.to_dict()

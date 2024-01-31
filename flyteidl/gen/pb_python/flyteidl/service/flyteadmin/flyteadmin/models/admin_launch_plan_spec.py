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


class AdminLaunchPlanSpec(object):
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
        'workflow_id': 'CoreIdentifier',
        'entity_metadata': 'AdminLaunchPlanMetadata',
        'default_inputs': 'CoreParameterMap',
        'fixed_inputs': 'CoreLiteralMap',
        'role': 'str',
        'labels': 'AdminLabels',
        'annotations': 'AdminAnnotations',
        'auth': 'AdminAuth',
        'auth_role': 'AdminAuthRole',
        'security_context': 'CoreSecurityContext',
        'quality_of_service': 'CoreQualityOfService',
        'raw_output_data_config': 'AdminRawOutputDataConfig',
        'max_parallelism': 'int',
        'interruptible': 'bool',
        'overwrite_cache': 'bool',
        'envs': 'AdminEnvs'
    }

    attribute_map = {
        'workflow_id': 'workflow_id',
        'entity_metadata': 'entity_metadata',
        'default_inputs': 'default_inputs',
        'fixed_inputs': 'fixed_inputs',
        'role': 'role',
        'labels': 'labels',
        'annotations': 'annotations',
        'auth': 'auth',
        'auth_role': 'auth_role',
        'security_context': 'security_context',
        'quality_of_service': 'quality_of_service',
        'raw_output_data_config': 'raw_output_data_config',
        'max_parallelism': 'max_parallelism',
        'interruptible': 'interruptible',
        'overwrite_cache': 'overwrite_cache',
        'envs': 'envs'
    }

    def __init__(self, workflow_id=None, entity_metadata=None, default_inputs=None, fixed_inputs=None, role=None, labels=None, annotations=None, auth=None, auth_role=None, security_context=None, quality_of_service=None, raw_output_data_config=None, max_parallelism=None, interruptible=None, overwrite_cache=None, envs=None, _configuration=None):  # noqa: E501
        """AdminLaunchPlanSpec - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._workflow_id = None
        self._entity_metadata = None
        self._default_inputs = None
        self._fixed_inputs = None
        self._role = None
        self._labels = None
        self._annotations = None
        self._auth = None
        self._auth_role = None
        self._security_context = None
        self._quality_of_service = None
        self._raw_output_data_config = None
        self._max_parallelism = None
        self._interruptible = None
        self._overwrite_cache = None
        self._envs = None
        self.discriminator = None

        if workflow_id is not None:
            self.workflow_id = workflow_id
        if entity_metadata is not None:
            self.entity_metadata = entity_metadata
        if default_inputs is not None:
            self.default_inputs = default_inputs
        if fixed_inputs is not None:
            self.fixed_inputs = fixed_inputs
        if role is not None:
            self.role = role
        if labels is not None:
            self.labels = labels
        if annotations is not None:
            self.annotations = annotations
        if auth is not None:
            self.auth = auth
        if auth_role is not None:
            self.auth_role = auth_role
        if security_context is not None:
            self.security_context = security_context
        if quality_of_service is not None:
            self.quality_of_service = quality_of_service
        if raw_output_data_config is not None:
            self.raw_output_data_config = raw_output_data_config
        if max_parallelism is not None:
            self.max_parallelism = max_parallelism
        if interruptible is not None:
            self.interruptible = interruptible
        if overwrite_cache is not None:
            self.overwrite_cache = overwrite_cache
        if envs is not None:
            self.envs = envs

    @property
    def workflow_id(self):
        """Gets the workflow_id of this AdminLaunchPlanSpec.  # noqa: E501


        :return: The workflow_id of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: CoreIdentifier
        """
        return self._workflow_id

    @workflow_id.setter
    def workflow_id(self, workflow_id):
        """Sets the workflow_id of this AdminLaunchPlanSpec.


        :param workflow_id: The workflow_id of this AdminLaunchPlanSpec.  # noqa: E501
        :type: CoreIdentifier
        """

        self._workflow_id = workflow_id

    @property
    def entity_metadata(self):
        """Gets the entity_metadata of this AdminLaunchPlanSpec.  # noqa: E501


        :return: The entity_metadata of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: AdminLaunchPlanMetadata
        """
        return self._entity_metadata

    @entity_metadata.setter
    def entity_metadata(self, entity_metadata):
        """Sets the entity_metadata of this AdminLaunchPlanSpec.


        :param entity_metadata: The entity_metadata of this AdminLaunchPlanSpec.  # noqa: E501
        :type: AdminLaunchPlanMetadata
        """

        self._entity_metadata = entity_metadata

    @property
    def default_inputs(self):
        """Gets the default_inputs of this AdminLaunchPlanSpec.  # noqa: E501

        Input values to be passed for the execution. These can be overridden when an execution is created with this launch plan.  # noqa: E501

        :return: The default_inputs of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: CoreParameterMap
        """
        return self._default_inputs

    @default_inputs.setter
    def default_inputs(self, default_inputs):
        """Sets the default_inputs of this AdminLaunchPlanSpec.

        Input values to be passed for the execution. These can be overridden when an execution is created with this launch plan.  # noqa: E501

        :param default_inputs: The default_inputs of this AdminLaunchPlanSpec.  # noqa: E501
        :type: CoreParameterMap
        """

        self._default_inputs = default_inputs

    @property
    def fixed_inputs(self):
        """Gets the fixed_inputs of this AdminLaunchPlanSpec.  # noqa: E501

        Fixed, non-overridable inputs for the Launch Plan. These can not be overridden when an execution is created with this launch plan.  # noqa: E501

        :return: The fixed_inputs of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: CoreLiteralMap
        """
        return self._fixed_inputs

    @fixed_inputs.setter
    def fixed_inputs(self, fixed_inputs):
        """Sets the fixed_inputs of this AdminLaunchPlanSpec.

        Fixed, non-overridable inputs for the Launch Plan. These can not be overridden when an execution is created with this launch plan.  # noqa: E501

        :param fixed_inputs: The fixed_inputs of this AdminLaunchPlanSpec.  # noqa: E501
        :type: CoreLiteralMap
        """

        self._fixed_inputs = fixed_inputs

    @property
    def role(self):
        """Gets the role of this AdminLaunchPlanSpec.  # noqa: E501


        :return: The role of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: str
        """
        return self._role

    @role.setter
    def role(self, role):
        """Sets the role of this AdminLaunchPlanSpec.


        :param role: The role of this AdminLaunchPlanSpec.  # noqa: E501
        :type: str
        """

        self._role = role

    @property
    def labels(self):
        """Gets the labels of this AdminLaunchPlanSpec.  # noqa: E501

        Custom labels to be applied to the execution resource.  # noqa: E501

        :return: The labels of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: AdminLabels
        """
        return self._labels

    @labels.setter
    def labels(self, labels):
        """Sets the labels of this AdminLaunchPlanSpec.

        Custom labels to be applied to the execution resource.  # noqa: E501

        :param labels: The labels of this AdminLaunchPlanSpec.  # noqa: E501
        :type: AdminLabels
        """

        self._labels = labels

    @property
    def annotations(self):
        """Gets the annotations of this AdminLaunchPlanSpec.  # noqa: E501

        Custom annotations to be applied to the execution resource.  # noqa: E501

        :return: The annotations of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: AdminAnnotations
        """
        return self._annotations

    @annotations.setter
    def annotations(self, annotations):
        """Sets the annotations of this AdminLaunchPlanSpec.

        Custom annotations to be applied to the execution resource.  # noqa: E501

        :param annotations: The annotations of this AdminLaunchPlanSpec.  # noqa: E501
        :type: AdminAnnotations
        """

        self._annotations = annotations

    @property
    def auth(self):
        """Gets the auth of this AdminLaunchPlanSpec.  # noqa: E501

        Indicates the permission associated with workflow executions triggered with this launch plan.  # noqa: E501

        :return: The auth of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: AdminAuth
        """
        return self._auth

    @auth.setter
    def auth(self, auth):
        """Sets the auth of this AdminLaunchPlanSpec.

        Indicates the permission associated with workflow executions triggered with this launch plan.  # noqa: E501

        :param auth: The auth of this AdminLaunchPlanSpec.  # noqa: E501
        :type: AdminAuth
        """

        self._auth = auth

    @property
    def auth_role(self):
        """Gets the auth_role of this AdminLaunchPlanSpec.  # noqa: E501


        :return: The auth_role of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: AdminAuthRole
        """
        return self._auth_role

    @auth_role.setter
    def auth_role(self, auth_role):
        """Sets the auth_role of this AdminLaunchPlanSpec.


        :param auth_role: The auth_role of this AdminLaunchPlanSpec.  # noqa: E501
        :type: AdminAuthRole
        """

        self._auth_role = auth_role

    @property
    def security_context(self):
        """Gets the security_context of this AdminLaunchPlanSpec.  # noqa: E501


        :return: The security_context of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: CoreSecurityContext
        """
        return self._security_context

    @security_context.setter
    def security_context(self, security_context):
        """Sets the security_context of this AdminLaunchPlanSpec.


        :param security_context: The security_context of this AdminLaunchPlanSpec.  # noqa: E501
        :type: CoreSecurityContext
        """

        self._security_context = security_context

    @property
    def quality_of_service(self):
        """Gets the quality_of_service of this AdminLaunchPlanSpec.  # noqa: E501

        Indicates the runtime priority of the execution.  # noqa: E501

        :return: The quality_of_service of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: CoreQualityOfService
        """
        return self._quality_of_service

    @quality_of_service.setter
    def quality_of_service(self, quality_of_service):
        """Sets the quality_of_service of this AdminLaunchPlanSpec.

        Indicates the runtime priority of the execution.  # noqa: E501

        :param quality_of_service: The quality_of_service of this AdminLaunchPlanSpec.  # noqa: E501
        :type: CoreQualityOfService
        """

        self._quality_of_service = quality_of_service

    @property
    def raw_output_data_config(self):
        """Gets the raw_output_data_config of this AdminLaunchPlanSpec.  # noqa: E501

        Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.).  # noqa: E501

        :return: The raw_output_data_config of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: AdminRawOutputDataConfig
        """
        return self._raw_output_data_config

    @raw_output_data_config.setter
    def raw_output_data_config(self, raw_output_data_config):
        """Sets the raw_output_data_config of this AdminLaunchPlanSpec.

        Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.).  # noqa: E501

        :param raw_output_data_config: The raw_output_data_config of this AdminLaunchPlanSpec.  # noqa: E501
        :type: AdminRawOutputDataConfig
        """

        self._raw_output_data_config = raw_output_data_config

    @property
    def max_parallelism(self):
        """Gets the max_parallelism of this AdminLaunchPlanSpec.  # noqa: E501

        Controls the maximum number of tasknodes that can be run in parallel for the entire workflow. This is useful to achieve fairness. Note: MapTasks are regarded as one unit, and parallelism/concurrency of MapTasks is independent from this.  # noqa: E501

        :return: The max_parallelism of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: int
        """
        return self._max_parallelism

    @max_parallelism.setter
    def max_parallelism(self, max_parallelism):
        """Sets the max_parallelism of this AdminLaunchPlanSpec.

        Controls the maximum number of tasknodes that can be run in parallel for the entire workflow. This is useful to achieve fairness. Note: MapTasks are regarded as one unit, and parallelism/concurrency of MapTasks is independent from this.  # noqa: E501

        :param max_parallelism: The max_parallelism of this AdminLaunchPlanSpec.  # noqa: E501
        :type: int
        """

        self._max_parallelism = max_parallelism

    @property
    def interruptible(self):
        """Gets the interruptible of this AdminLaunchPlanSpec.  # noqa: E501

        Allows for the interruptible flag of a workflow to be overwritten for a single execution. Omitting this field uses the workflow's value as a default. As we need to distinguish between the field not being provided and its default value false, we have to use a wrapper around the bool field.  # noqa: E501

        :return: The interruptible of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: bool
        """
        return self._interruptible

    @interruptible.setter
    def interruptible(self, interruptible):
        """Sets the interruptible of this AdminLaunchPlanSpec.

        Allows for the interruptible flag of a workflow to be overwritten for a single execution. Omitting this field uses the workflow's value as a default. As we need to distinguish between the field not being provided and its default value false, we have to use a wrapper around the bool field.  # noqa: E501

        :param interruptible: The interruptible of this AdminLaunchPlanSpec.  # noqa: E501
        :type: bool
        """

        self._interruptible = interruptible

    @property
    def overwrite_cache(self):
        """Gets the overwrite_cache of this AdminLaunchPlanSpec.  # noqa: E501

        Allows for all cached values of a workflow and its tasks to be overwritten for a single execution. If enabled, all calculations are performed even if cached results would be available, overwriting the stored data once execution finishes successfully.  # noqa: E501

        :return: The overwrite_cache of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: bool
        """
        return self._overwrite_cache

    @overwrite_cache.setter
    def overwrite_cache(self, overwrite_cache):
        """Sets the overwrite_cache of this AdminLaunchPlanSpec.

        Allows for all cached values of a workflow and its tasks to be overwritten for a single execution. If enabled, all calculations are performed even if cached results would be available, overwriting the stored data once execution finishes successfully.  # noqa: E501

        :param overwrite_cache: The overwrite_cache of this AdminLaunchPlanSpec.  # noqa: E501
        :type: bool
        """

        self._overwrite_cache = overwrite_cache

    @property
    def envs(self):
        """Gets the envs of this AdminLaunchPlanSpec.  # noqa: E501

        Environment variables to be set for the execution.  # noqa: E501

        :return: The envs of this AdminLaunchPlanSpec.  # noqa: E501
        :rtype: AdminEnvs
        """
        return self._envs

    @envs.setter
    def envs(self, envs):
        """Sets the envs of this AdminLaunchPlanSpec.

        Environment variables to be set for the execution.  # noqa: E501

        :param envs: The envs of this AdminLaunchPlanSpec.  # noqa: E501
        :type: AdminEnvs
        """

        self._envs = envs

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
        if issubclass(AdminLaunchPlanSpec, dict):
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
        if not isinstance(other, AdminLaunchPlanSpec):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, AdminLaunchPlanSpec):
            return True

        return self.to_dict() != other.to_dict()

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


class EventWorkflowExecutionEvent(object):
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
        'execution_id': 'CoreWorkflowExecutionIdentifier',
        'producer_id': 'str',
        'phase': 'CoreWorkflowExecutionPhase',
        'occurred_at': 'datetime',
        'output_uri': 'str',
        'error': 'CoreExecutionError',
        'output_data': 'CoreLiteralMap'
    }

    attribute_map = {
        'execution_id': 'execution_id',
        'producer_id': 'producer_id',
        'phase': 'phase',
        'occurred_at': 'occurred_at',
        'output_uri': 'output_uri',
        'error': 'error',
        'output_data': 'output_data'
    }

    def __init__(self, execution_id=None, producer_id=None, phase=None, occurred_at=None, output_uri=None, error=None, output_data=None, _configuration=None):  # noqa: E501
        """EventWorkflowExecutionEvent - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._execution_id = None
        self._producer_id = None
        self._phase = None
        self._occurred_at = None
        self._output_uri = None
        self._error = None
        self._output_data = None
        self.discriminator = None

        if execution_id is not None:
            self.execution_id = execution_id
        if producer_id is not None:
            self.producer_id = producer_id
        if phase is not None:
            self.phase = phase
        if occurred_at is not None:
            self.occurred_at = occurred_at
        if output_uri is not None:
            self.output_uri = output_uri
        if error is not None:
            self.error = error
        if output_data is not None:
            self.output_data = output_data

    @property
    def execution_id(self):
        """Gets the execution_id of this EventWorkflowExecutionEvent.  # noqa: E501


        :return: The execution_id of this EventWorkflowExecutionEvent.  # noqa: E501
        :rtype: CoreWorkflowExecutionIdentifier
        """
        return self._execution_id

    @execution_id.setter
    def execution_id(self, execution_id):
        """Sets the execution_id of this EventWorkflowExecutionEvent.


        :param execution_id: The execution_id of this EventWorkflowExecutionEvent.  # noqa: E501
        :type: CoreWorkflowExecutionIdentifier
        """

        self._execution_id = execution_id

    @property
    def producer_id(self):
        """Gets the producer_id of this EventWorkflowExecutionEvent.  # noqa: E501


        :return: The producer_id of this EventWorkflowExecutionEvent.  # noqa: E501
        :rtype: str
        """
        return self._producer_id

    @producer_id.setter
    def producer_id(self, producer_id):
        """Sets the producer_id of this EventWorkflowExecutionEvent.


        :param producer_id: The producer_id of this EventWorkflowExecutionEvent.  # noqa: E501
        :type: str
        """

        self._producer_id = producer_id

    @property
    def phase(self):
        """Gets the phase of this EventWorkflowExecutionEvent.  # noqa: E501


        :return: The phase of this EventWorkflowExecutionEvent.  # noqa: E501
        :rtype: CoreWorkflowExecutionPhase
        """
        return self._phase

    @phase.setter
    def phase(self, phase):
        """Sets the phase of this EventWorkflowExecutionEvent.


        :param phase: The phase of this EventWorkflowExecutionEvent.  # noqa: E501
        :type: CoreWorkflowExecutionPhase
        """

        self._phase = phase

    @property
    def occurred_at(self):
        """Gets the occurred_at of this EventWorkflowExecutionEvent.  # noqa: E501

        This timestamp represents when the original event occurred, it is generated by the executor of the workflow.  # noqa: E501

        :return: The occurred_at of this EventWorkflowExecutionEvent.  # noqa: E501
        :rtype: datetime
        """
        return self._occurred_at

    @occurred_at.setter
    def occurred_at(self, occurred_at):
        """Sets the occurred_at of this EventWorkflowExecutionEvent.

        This timestamp represents when the original event occurred, it is generated by the executor of the workflow.  # noqa: E501

        :param occurred_at: The occurred_at of this EventWorkflowExecutionEvent.  # noqa: E501
        :type: datetime
        """

        self._occurred_at = occurred_at

    @property
    def output_uri(self):
        """Gets the output_uri of this EventWorkflowExecutionEvent.  # noqa: E501

        URL to the output of the execution, it encodes all the information including Cloud source provider. ie., s3://...  # noqa: E501

        :return: The output_uri of this EventWorkflowExecutionEvent.  # noqa: E501
        :rtype: str
        """
        return self._output_uri

    @output_uri.setter
    def output_uri(self, output_uri):
        """Sets the output_uri of this EventWorkflowExecutionEvent.

        URL to the output of the execution, it encodes all the information including Cloud source provider. ie., s3://...  # noqa: E501

        :param output_uri: The output_uri of this EventWorkflowExecutionEvent.  # noqa: E501
        :type: str
        """

        self._output_uri = output_uri

    @property
    def error(self):
        """Gets the error of this EventWorkflowExecutionEvent.  # noqa: E501


        :return: The error of this EventWorkflowExecutionEvent.  # noqa: E501
        :rtype: CoreExecutionError
        """
        return self._error

    @error.setter
    def error(self, error):
        """Sets the error of this EventWorkflowExecutionEvent.


        :param error: The error of this EventWorkflowExecutionEvent.  # noqa: E501
        :type: CoreExecutionError
        """

        self._error = error

    @property
    def output_data(self):
        """Gets the output_data of this EventWorkflowExecutionEvent.  # noqa: E501

        Raw output data produced by this workflow execution.  # noqa: E501

        :return: The output_data of this EventWorkflowExecutionEvent.  # noqa: E501
        :rtype: CoreLiteralMap
        """
        return self._output_data

    @output_data.setter
    def output_data(self, output_data):
        """Sets the output_data of this EventWorkflowExecutionEvent.

        Raw output data produced by this workflow execution.  # noqa: E501

        :param output_data: The output_data of this EventWorkflowExecutionEvent.  # noqa: E501
        :type: CoreLiteralMap
        """

        self._output_data = output_data

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
        if issubclass(EventWorkflowExecutionEvent, dict):
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
        if not isinstance(other, EventWorkflowExecutionEvent):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, EventWorkflowExecutionEvent):
            return True

        return self.to_dict() != other.to_dict()

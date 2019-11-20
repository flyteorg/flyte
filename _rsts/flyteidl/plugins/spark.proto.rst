.. _api_file_flyteidl/plugins/spark.proto:

spark.proto
============================

.. _api_msg_flyteidl.plugins.SparkApplication:

flyteidl.plugins.SparkApplication
---------------------------------

`[flyteidl.plugins.SparkApplication proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/spark.proto#L6>`_


.. code-block:: json

  {}



.. _api_enum_flyteidl.plugins.SparkApplication.Type:

Enum flyteidl.plugins.SparkApplication.Type
-------------------------------------------

`[flyteidl.plugins.SparkApplication.Type proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/spark.proto#L7>`_


.. _api_enum_value_flyteidl.plugins.SparkApplication.Type.PYTHON:

PYTHON
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.plugins.SparkApplication.Type.JAVA:

JAVA
  ⁣
  
.. _api_enum_value_flyteidl.plugins.SparkApplication.Type.SCALA:

SCALA
  ⁣
  
.. _api_enum_value_flyteidl.plugins.SparkApplication.Type.R:

R
  ⁣
  

.. _api_msg_flyteidl.plugins.SparkJob:

flyteidl.plugins.SparkJob
-------------------------

`[flyteidl.plugins.SparkJob proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/spark.proto#L16>`_

Custom Proto for Spark Plugin.

.. code-block:: json

  {
    "applicationType": "...",
    "mainApplicationFile": "...",
    "mainClass": "...",
    "sparkConf": "{...}",
    "hadoopConf": "{...}",
    "executorPath": "..."
  }

.. _api_field_flyteidl.plugins.SparkJob.applicationType:

applicationType
  (:ref:`flyteidl.plugins.SparkApplication.Type <api_enum_flyteidl.plugins.SparkApplication.Type>`) 
  
.. _api_field_flyteidl.plugins.SparkJob.mainApplicationFile:

mainApplicationFile
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.plugins.SparkJob.mainClass:

mainClass
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.plugins.SparkJob.sparkConf:

sparkConf
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, `string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_>) 
  
.. _api_field_flyteidl.plugins.SparkJob.hadoopConf:

hadoopConf
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, `string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_>) 
  
.. _api_field_flyteidl.plugins.SparkJob.executorPath:

executorPath
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  



.. _api_file_flyteidl/core/condition.proto:

condition.proto
=============================

.. _api_msg_flyteidl.core.ComparisonExpression:

flyteidl.core.ComparisonExpression
----------------------------------

`[flyteidl.core.ComparisonExpression proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/condition.proto#L10>`_

Defines a 2-level tree where the root is a comparison operator and Operands are primitives or known variables.
Each expression results in a boolean result.

.. code-block:: json

  {
    "operator": "...",
    "left_value": "{...}",
    "right_value": "{...}"
  }

.. _api_field_flyteidl.core.ComparisonExpression.operator:

operator
  (:ref:`flyteidl.core.ComparisonExpression.Operator <api_enum_flyteidl.core.ComparisonExpression.Operator>`) 
  
.. _api_field_flyteidl.core.ComparisonExpression.left_value:

left_value
  (:ref:`flyteidl.core.Operand <api_msg_flyteidl.core.Operand>`) 
  
.. _api_field_flyteidl.core.ComparisonExpression.right_value:

right_value
  (:ref:`flyteidl.core.Operand <api_msg_flyteidl.core.Operand>`) 
  

.. _api_enum_flyteidl.core.ComparisonExpression.Operator:

Enum flyteidl.core.ComparisonExpression.Operator
------------------------------------------------

`[flyteidl.core.ComparisonExpression.Operator proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/condition.proto#L12>`_

Binary Operator for each expression

.. _api_enum_value_flyteidl.core.ComparisonExpression.Operator.EQ:

EQ
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.ComparisonExpression.Operator.NEQ:

NEQ
  ⁣
  
.. _api_enum_value_flyteidl.core.ComparisonExpression.Operator.GT:

GT
  ⁣Greater Than
  
  
.. _api_enum_value_flyteidl.core.ComparisonExpression.Operator.GTE:

GTE
  ⁣
  
.. _api_enum_value_flyteidl.core.ComparisonExpression.Operator.LT:

LT
  ⁣Less Than
  
  
.. _api_enum_value_flyteidl.core.ComparisonExpression.Operator.LTE:

LTE
  ⁣
  

.. _api_msg_flyteidl.core.Operand:

flyteidl.core.Operand
---------------------

`[flyteidl.core.Operand proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/condition.proto#L29>`_

Defines an operand to a comparison expression.

.. code-block:: json

  {
    "primitive": "{...}",
    "var": "..."
  }

.. _api_field_flyteidl.core.Operand.primitive:

primitive
  (:ref:`flyteidl.core.Primitive <api_msg_flyteidl.core.Primitive>`) Can be a constant
  
  
  
  Only one of :ref:`primitive <api_field_flyteidl.core.Operand.primitive>`, :ref:`var <api_field_flyteidl.core.Operand.var>` may be set.
  
.. _api_field_flyteidl.core.Operand.var:

var
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Or one of this node's input variables
  
  
  
  Only one of :ref:`primitive <api_field_flyteidl.core.Operand.primitive>`, :ref:`var <api_field_flyteidl.core.Operand.var>` may be set.
  


.. _api_msg_flyteidl.core.BooleanExpression:

flyteidl.core.BooleanExpression
-------------------------------

`[flyteidl.core.BooleanExpression proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/condition.proto#L40>`_

Defines a boolean expression tree. It can be a simple or a conjunction expression.
Multiple expressions can be combined using a conjunction or a disjunction to result in a final boolean result.

.. code-block:: json

  {
    "conjunction": "{...}",
    "comparison": "{...}"
  }

.. _api_field_flyteidl.core.BooleanExpression.conjunction:

conjunction
  (:ref:`flyteidl.core.ConjunctionExpression <api_msg_flyteidl.core.ConjunctionExpression>`) 
  
  
  Only one of :ref:`conjunction <api_field_flyteidl.core.BooleanExpression.conjunction>`, :ref:`comparison <api_field_flyteidl.core.BooleanExpression.comparison>` may be set.
  
.. _api_field_flyteidl.core.BooleanExpression.comparison:

comparison
  (:ref:`flyteidl.core.ComparisonExpression <api_msg_flyteidl.core.ComparisonExpression>`) 
  
  
  Only one of :ref:`conjunction <api_field_flyteidl.core.BooleanExpression.conjunction>`, :ref:`comparison <api_field_flyteidl.core.BooleanExpression.comparison>` may be set.
  


.. _api_msg_flyteidl.core.ConjunctionExpression:

flyteidl.core.ConjunctionExpression
-----------------------------------

`[flyteidl.core.ConjunctionExpression proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/condition.proto#L48>`_

Defines a conjunction expression of two boolean expressions.

.. code-block:: json

  {
    "operator": "...",
    "left_expression": "{...}",
    "right_expression": "{...}"
  }

.. _api_field_flyteidl.core.ConjunctionExpression.operator:

operator
  (:ref:`flyteidl.core.ConjunctionExpression.LogicalOperator <api_enum_flyteidl.core.ConjunctionExpression.LogicalOperator>`) 
  
.. _api_field_flyteidl.core.ConjunctionExpression.left_expression:

left_expression
  (:ref:`flyteidl.core.BooleanExpression <api_msg_flyteidl.core.BooleanExpression>`) 
  
.. _api_field_flyteidl.core.ConjunctionExpression.right_expression:

right_expression
  (:ref:`flyteidl.core.BooleanExpression <api_msg_flyteidl.core.BooleanExpression>`) 
  

.. _api_enum_flyteidl.core.ConjunctionExpression.LogicalOperator:

Enum flyteidl.core.ConjunctionExpression.LogicalOperator
--------------------------------------------------------

`[flyteidl.core.ConjunctionExpression.LogicalOperator proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/condition.proto#L51>`_

Nested conditions. They can be conjoined using AND / OR
Order of evaluation is not important as the operators are Commutative

.. _api_enum_value_flyteidl.core.ConjunctionExpression.LogicalOperator.AND:

AND
  *(DEFAULT)* ⁣Conjunction
  
  
.. _api_enum_value_flyteidl.core.ConjunctionExpression.LogicalOperator.OR:

OR
  ⁣
  

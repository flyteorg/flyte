{{ fullname | escape | underline}}

.. currentmodule:: {{ module }}

{% if objname == 'FlyteFile' or objname == 'FlyteDirectory' %}

.. autoclass:: {{ objname }}

   {% block methods %}
   {% if methods %}

   .. rubric:: {{ _('Methods') }}
   {% for item in methods %}

   {% if item != '__init__' %}
   .. automethod:: {{ item }}
   {% endif %}

   {%- endfor %}
   {% endif %}
   {% endblock %}

   {% block attributes %}
   {% if attributes %}

   .. rubric:: {{ _('Attributes') }}
   {% for item in attributes %}
   .. autoattribute:: {{ item }}
   {%- endfor %}

   {% endif %}
   {% endblock %}


{% else %}

.. autodata:: {{ objname }}

{% endif %}

{{ fullname | escape | underline}}

.. currentmodule:: {{ module }}

{% if objtype == 'class' %}

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
      :noindex:
   {%- endfor %}

   {% endif %}
   {% endblock %}


{% endif %}

{% if objtype == 'function' %}

.. autofunction:: {{ objname }}

{% endif %}

.. _divedeep-projects:

Projects
========

.. tags:: Basic, Glossary

A project in Flyte is a group of :ref:`workflows <divedeep-workflows>` and :ref:`tasks <divedeep-tasks>` tied together to achieve a goal. 

A Flyte project can map to an engineering project or everything that's owned by a team or an individual. There cannot be multiple projects with the same name in Flyte. 

Since the fully-qualified name for tasks and workflows include the project and domain name, the task/workflow names are only required to be unique within a project. The workflows in a project ``A`` can refer to tasks and workflows in other projects using the fully-qualified name.

Flyte allows users to set resource limits and provides basic reports and dashboards automatically for each project. The information captured in these reports includes workflow/task level insights, resource usage, and billing information.
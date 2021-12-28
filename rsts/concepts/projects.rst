.. _divedeep-projects:

Projects
========
A project in Flyte is a grouping of workflows and tasks to achieve a particular goal. A Flyte project can map to an engineering project or everything that's owned by a team or an individual. There cannot be multiple projects with the same name in Flyte. 

Since the fully-qualified name for tasks and workflows include project name and domain name, the task/workflow names are only required to be unique within a project. The workflows in a project A can refer tasks and workflows from other projects using the fully-qualified name.

Flyte allows users to set resource limit for each project. Flyte provides basic reports and dashboards automatically for each project. The information captured in these reports include workflow/task level insights, resource usage and billing information.
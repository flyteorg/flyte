# -*- coding: utf-8 -*-
#
# Configuration file for the Sphinx documentation builder.
#
# This file does only contain a selection of the most common options. For a
# full list see the documentation:
# http://www.sphinx-doc.org/en/stable/config

# -- Path setup --------------------------------------------------------------

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use pathlib.Path.resolve(strict=True) to make it absolute, like shown here.

import os
from pathlib import Path
import logging
import sys
import inspect

import sphinx.application
import sphinx.errors
from sphinx.util import logging as sphinx_logging


sys.path.insert(0, str(Path("../").resolve(strict=True)))
sys.path.append(str(Path("./_ext").resolve(strict=True)))

sphinx.application.ExtensionError = sphinx.errors.ExtensionError

# -- Project information -----------------------------------------------------

project = "Flyte"
copyright = "2024, Flyte authors"
author = "Flyte"

# The short X.Y version
version = ""
# The full version, including alpha/beta/rc tags
release = "1.13.2"

# -- General configuration ---------------------------------------------------

# If your documentation needs a minimal Sphinx version, state it here.
#
# needs_sphinx = '1.0'

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
extensions = [
    "sphinx.ext.napoleon",
    "sphinx.ext.autodoc",
    "sphinx.ext.autosummary",
    "sphinx.ext.autosectionlabel",
    "sphinx.ext.doctest",
    "sphinx.ext.inheritance_diagram",
    "sphinx.ext.intersphinx",
    "sphinx.ext.graphviz",
    "sphinx.ext.todo",
    "sphinx.ext.coverage",
    "sphinx.ext.linkcode",
    "sphinx.ext.ifconfig",
    "sphinx.ext.extlinks",
    "sphinx-prompt",
    "sphinx_copybutton",
    "sphinx_docsearch",
    "sphinxext.remoteliteralinclude",
    "sphinx_issues",
    "sphinx_click",
    "sphinx_design",
    "sphinx_reredirects",
    "sphinxcontrib.mermaid",
    "sphinxcontrib.video",
    "sphinxcontrib.youtube",
    "sphinx_tabs.tabs",
    "sphinx_tags",
    "myst_nb",
    # custom extensions
    "auto_examples",
    "import_projects",
    "notebook_literalinclude",
]

source_suffix = {
    ".rst": "restructuredtext",
    ".md": "myst-nb",
}

extlinks = {
    "propeller": ("https://github.com/flyteorg/flytepropeller/tree/master/%s", ""),
    "stdlib": ("https://github.com/flyteorg/flytestdlib/tree/master/%s", ""),
    "kit": ("https://github.com/flyteorg/flytekit/tree/master/%s", ""),
    "plugins": ("https://github.com/flyteorg/flyteplugins/tree/v0.1.4/%s", ""),
    "idl": ("https://github.com/flyteorg/flyteidl/tree/v0.14.1/%s", ""),
    "admin": ("https://github.com/flyteorg/flyteadmin/tree/master/%s", ""),
    "cookbook": ("https://flytecookbook.readthedocs.io/en/latest/", None),
}

# redirects
redirects = {
    # concepts
    "concepts/admin": "../user_guide/concepts/control_plane/admin.html",
    "concepts/catalog": "../user_guide/concepts/main_concepts/catalog.html",
    "concepts/console": "../user_guide/concepts/control_plane/console.html",
    "concepts/data_management": "../user_guide/concepts/main_concepts/data_management.html",
    "concepts/domains": "../user_guide/concepts/control_plane/domains.html",
    "concepts/dynamic_spec": "../user_guide/concepts/control_plane/dynamic_spec.html",
    "concepts/execution_timeline": "../user_guide/concepts/execution_timeline.html",
    "concepts/executions": "../user_guide/concepts/executions.html",
    "concepts/flyte_console": "../user_guide/concepts/main_concepts/flyte_console.html",
    "concepts/flytepropeller_architecture": "../user_guide/concepts/component_architecture/flytepropeller_architecture.html",
    "concepts/launchplans": "../user_guide/concepts/main_concepts/launchplans.html",
    "concepts/native_scheduler": "../user_guide/concepts/component_architecture/native_scheduler.html",
    "concepts/nodes": "../user_guide/concepts/main_concepts/nodes.html",
    "concepts/projects": "../user_guide/concepts/control_plane/projects.html",
    "concepts/registration": "../user_guide/concepts/main_concepts/registration.html",
    "concepts/schedules": "../user_guide/concepts/main_concepts/schedules.html",
    "concepts/state_machine": "../user_guide/concepts/main_concepts/state_machine.html",
    "concepts/tasks": "../user_guide/concepts/main_concepts/tasks.html",
    "concepts/versioning": "../user_guide/concepts/main_concepts/versioning.html",
    "concepts/workflow_lifecycle": "../user_guide/concepts/main_concepts/workflow_lifecycle.html",
    "concepts/workflows": "../user_guide/concepts/main_concepts/workflows.html",

    # core use cases
    "core_use_cases/index": "../flytesnacks/tutorials/index.html",
    "core_use_cases/analytics": "../flytesnacks/tutorials/index.html",
    "core_use_cases/data_engineering": "../flytesnacks/tutorials/index.html",
    "core_use_cases/machine_learning": "../flytesnacks/tutorials/index.html",

    # deprecated integrations
    "deprecated_integrations/index": "../../flytesnacks/deprecated_integrations/index.html",
    "deprecated_integrations/mmcloud_plugin/index": "../../flytesnacks/examples/mmcloud_agent/index.html",
    "deprecated_integrations/mmcloud_plugin/mmcloud_plugin_example": "../../flytesnacks/examples/mmcloud_agent/index.html",
    "deprecated_integrations/bigquery_plugin/index": "../../flytesnacks/deprecated_integrations/bigquery_plugin/index.html",
    "deprecated_integrations/bigquery_plugin/bigquery_plugin_example": "../../flytesnacks/deprecated_integrations/bigquery_plugin/bigquery_plugin_example.html",
    "deprecated_integrations/databricks_plugin/index": "../../flytesnacks/deprecated_integrations/databricks_plugin/index.html",
    "deprecated_integrations/databricks_plugin/databricks_plugin_example": "../../flytesnacks/deprecated_integrations/databricks_plugin/databricks_plugin_example.html",
    "deprecated_integrations/snowflake_plugin/index": "../../flytesnacks/deprecated_integrations/snowflake_plugin/index.html",
    "deprecated_integrations/snowflake_plugin/snowflake_plugin_example": "../../flytesnacks/deprecated_integrations/snowflake_plugin/snowflake_plugin_example.html",

    # flyte agents
    "flyte_agents/index": "../user_guide/flyte_agents/index.html",
    "flyte_agents/deploying_agents_to_the_flyte_sandbox": "../user_guide/flyte_agents/deploying_agents_to_the_flyte_sandbox.html",
    "flyte_agents/enabling_agents_in_your_flyte_deployment": "../user_guide/flyte_agents/enabling_agents_in_your_flyte_deployment.html",
    "flyte_agents/how_secret_works_in_agent": "../user_guide/flyte_agents/how_secret_works_in_agent.html",
    "flyte_agents/implementing_the_agent_metadata_service": "../user_guide/flyte_agents/implementing_the_agent_metadata_service.html",
    "flyte_agents/testing_agents_in_a_local_development_cluster": "../user_guide/flyte_agents/testing_agents_in_a_local_development_cluster.html",
    "flyte_agents/testing_agents_in_a_local_python_environment": "../user_guide/flyte_agents/testing_agents_in_a_local_python_environment.html",

    # flyte fundamentals
    "flyte_fundamentals/index": "../user_guide/flyte_fundamentals/index.html",
    "flyte_fundamentals/extending_flyte": "../user_guide/flyte_fundamentals/extending_flyte.html",
    "flyte_fundamentals/optimizing_tasks": "../user_guide/flyte_fundamentals/optimizing_tasks.html",
    "flyte_fundamentals/registering_workflows": "../user_guide/flyte_fundamentals/registering_workflows.html",
    "flyte_fundamentals/running_and_scheduling_workflows": "../user_guide/flyte_fundamentals/running_and_scheduling_workflows.html",
    "flyte_fundamentals/tasks_workflows_and_launch_plans": "../user_guide/flyte_fundamentals/tasks_workflows_and_launch_plans.html",
    "flyte_fundamentals/visualizing_task_input_and_output": "../user_guide/flyte_fundamentals/visualizing_task_input_and_output.html",

    # flytesnacks
    "flytesnacks/contribute": "../community/contribute/contribute_docs.html",
    "community/contribute": "community/contribute/contribute_code.html",
    "flytesnacks/integrations": "../flytesnacks/integrations/index.html",
    "flytesnacks/tutorials": "../flytesnacks/tutorials.html",

    # getting started with workflow development
    "getting_started_with_workflow_development/index": "../user_guide/getting_started_with_workflow_development/index.html",
    "getting_started_with_workflow_development/creating_a_flyte_project": "../user_guide/getting_started_with_workflow_development/creating_a_flyte_project.html",
    "getting_started_with_workflow_development/flyte_project_components": "../user_guide/getting_started_with_workflow_development/flyte_project_components.html",
    "getting_started_with_workflow_development/installing_development_tools": "../user_guide/getting_started_with_workflow_development/installing_development_tools.html",
    "getting_started_with_workflow_development/running_a_workflow_locally": "../user_guide/getting_started_with_workflow_development/running_a_workflow_locally.html",

    # misc standalone pages
    "introduction": "../user_guide/introduction.html",
    "quickstart_guide": "../user_guide/quickstart_guide.html",

    # flytectl
    "flytectl/docs_index": "../../api/flytectl/docs_index.html",
    "flytectl/overview": "../../api/flytectl/overview.html",
    "flytectl/gen/flytectl": "../../api/flytectl/gen/flytectl.html",

    # flytectl verbs
    "flytectl/verbs": "../../api/flytectl/verbs.html",
    "flytectl/gen/flytectl_create": "../../api/flytectl/gen/flytectl_create.html",
    "flytectl/gen/flytectl_completion": "../../api/flytectl/gen/flytectl_completion.html",
    "flytectl/gen/flytectl_get": "../../api/flytectl/gen/flytectl_get.html",
    "flytectl/gen/flytectl_update": "../../api/flytectl/gen/flytectl_update.html",
    "flytectl/gen/flytectl_delete": "../../api/flytectl/gen/flytectl_delete.html",
    "flytectl/gen/flytectl_register": "../../api/flytectl/gen/flytectl_register.html",
    "flytectl/gen/flytectl_config": "../../api/flytectl/gen/flytectl_config.html",
    "flytectl/gen/flytectl_compile": "../../api/flytectl/gen/flytectl_compile.html",
    "flytectl/gen/flytectl_sandbox": "../../api/flytectl/gen/flytectl_sandbox.html",
    "flytectl/gen/flytectl_demo": "../../api/flytectl/gen/flytectl_demo.html",
    "flytectl/gen/flytectl_version": "../../api/flytectl/gen/flytectl_version.html",
    "flytectl/gen/flytectl_upgrade": "../../api/flytectl/gen/flytectl_upgrade.html",

    # flytectl nouns
    "flytectl/nouns": "../../api/flytectl/nouns.html",

    # flytectl project
    "flytectl/project": "../../api/flytectl/project.html",
    "flytectl/gen/flytectl_create_project": "../../api/flytectl/gen/flytectl_create_project.html",
    "flytectl/gen/flytectl_create_project": "../../api/flytectl/gen/flytectl_create_project.html",
    "flytectl/gen/flytectl_update_project": "../../api/flytectl/gen/flytectl_update_project.html",

    # flytectl execution
    "flytectl/execution": "../../api/flytectl/execution.html",
    "flytectl/gen/flytectl_create_execution": "../../api/flytectl/gen/flytectl_create_execution.html",
    "flytectl/gen/flytectl_get_execution": "../../api/flytectl/gen/flytectl_get_execution.html",
    "flytectl/gen/flytectl_update_execution": "../../api/flytectl/gen/flytectl_update_execution.html",
    "flytectl/gen/flytectl_delete_execution": "../../api/flytectl/gen/flytectl_delete_execution.html",

    # flytectl workflow
    "flytectl/workflow": "../../api/flytectl/workflow.html",
    "flytectl/gen/flytectl_get_workflow": "../../api/flytectl/gen/flytectl_get_workflow.html",
    "flytectl/gen/flytectl_update_workflow-meta": "../../api/flytectl/gen/flytectl_update_workflow-meta.html",

    # flytectl task
    "flytectl/task": "../../api/flytectl/task.html",
    "flytectl/gen/flytectl_get_task": "../../api/flytectl/gen/flytectl_get_task.html",
    "flytectl/gen/flytectl_update_task-meta": "../../api/flytectl/gen/flytectl_update_task-meta.html",

    # flytectl task resource attribute
    "flytectl/task-resource-attribute": "../../api/flytectl/task-resource-attribute.html",
    "flytectl/gen/flytectl_get_task-resource-attribute": "../../api/flytectl/gen/flytectl_get_task-resource-attribute.html",
    "flytectl/gen/flytectl_update_task-resource-attribute": "../../api/flytectl/gen/flytectl_update_task-resource-attribute.html",
    "flytectl/gen/flytectl_delete_task-resource-attribute": "../../api/flytectl/gen/flytectl_delete_task-resource-attribute.html",

    # flytectl cluster resource attribute
    "flytectl/cluster-resource-attribute": "../../api/flytectl/cluster-resource-attribute.html",
    "flytectl/gen/flytectl_get_cluster-resource-attribute": "../../api/flytectl/gen/flytectl_get_cluster-resource-attribute.html",
    "flytectl/gen/flytectl_update_cluster-resource-attribute": "../../api/flytectl/gen/flytectl_update_cluster-resource-attribute.html",
    "flytectl/gen/flytectl_delete_cluster-resource-attribute": "../../api/flytectl/gen/flytectl_delete_cluster-resource-attribute.html",

    # flytectl execution cluster label
    "flytectl/execution-cluster-label": "../../api/flytectl/execution-cluster-label.html",
    "flytectl/gen/flytectl_get_execution-cluster-label": "../../api/flytectl/gen/flytectl_get_execution-cluster-label.html",
    "flytectl/gen/flytectl_update_execution-cluster-label": "../../api/flytectl/gen/flytectl_update_execution-cluster-label.html",
    "flytectl/gen/flytectl_delete_execution-cluster-label": "../../api/flytectl/gen/flytectl_delete_execution-cluster-label",

    # flytectl execution queue attribute
    "flytectl/execution-queue-attribute": "../../api/flytectl/execution-queue-attribute.html",
    "flytectl/gen/flytectl_get_execution-queue-attribute": "../../api/flytectl/gen/flytectl_get_execution-queue-attribute.html",
    "flytectl/gen/flytectl_update_execution-queue-attribute": "../../api/flytectl/gen/flytectl_update_execution-queue-attribute.html",
    "flytectl/gen/flytectl_delete_execution-queue-attribute": "../../api/flytectl/gen/flytectl_delete_execution-queue-attribute",

    # flytectl plugin override
    "flytectl/plugin-override": "../../api/flytectl/plugin-override.html",
    "flytectl/gen/flytectl_get_plugin-override": "../../api/flytectl/gen/flytectl_get_plugin-override.html",
    "flytectl/gen/flytectl_update_plugin-override": "../../api/flytectl/gen/flytectl_update_plugin-override.html",
    "flytectl/gen/flytectl_delete_plugin-override": "../../api/flytectl/gen/flytectl_delete_plugin-override",

    # flytectl launchplan
    "flytectl/launchplan": "../../api/flytectl/launchplan.html",
    "flytectl/gen/flytectl_get_launchplan": "../../api/flytectl/gen/flytectl_get_launchplan.html",
    "flytectl/gen/flytectl_update_launchplan": "../../api/flytectl/gen/flytectl_update_launchplan.html",
    "flytectl/gen/flytectl_update_launchplan-meta": "../../api/flytectl/gen/flytectl_update_launchplan-meta.html",

    # flytectl workflow execution config
    "flytectl/workflow-execution-config": "../../api/flytectl/workflow-execution-config.html",
    "flytectl/gen/flytectl_get_workflow-execution-config": "../../api/flytectl/gen/flytectl_get_workflow-execution-config.html",
    "flytectl/gen/flytectl_update_workflow-execution-config": "../../api/flytectl/gen/flytectl_update_workflow-execution-config.html",
    "flytectl/gen/flytectl_delete_workflow-execution-config": "../../api/flytectl/gen/flytectl_delete_workflow-execution-config",

    # flytectl examples
    "flytectl/examples": "../../api/flytectl/examples.html",
    "flytectl/gen/flytectl_register_examples": "../../api/flytectl/gen/flytectl_register_examples",

    # flytectl files
    "flytectl/files": "../../api/flytectl/files.html",
    "flytectl/gen/flytectl_register_files": "../../api/flytectl/gen/flytectl_register_files",

    # flytectl config
    "flytectl/config": "../../api/flytectl/config.html",
    "flytectl/gen/flytectl_config_validate": "../../api/flytectl/gen/flytectl_config_validate",
    "flytectl/gen/flytectl_config_init": "../../api/flytectl/gen/flytectl_config_init",
    "flytectl/gen/flytectl_config_docs": "../../api/flytectl/gen/flytectl_config_docs",
    "flytectl/gen/flytectl_config_discover": "../../api/flytectl/gen/flytectl_config_discover",

    # flytectl sandbox
    "flytectl/sandbox": "../../api/flytectl/sandbox.html",
    "flytectl/gen/flytectl_sandbox_start": "../../api/flytectl/gen/flytectl_sandbox_start",
    "flytectl/gen/flytectl_sandbox_status": "../../api/flytectl/gen/flytectl_sandbox_status",
    "flytectl/gen/flytectl_sandbox_teardown": "../../api/flytectl/gen/flytectl_sandbox_teardown",
    "flytectl/gen/flytectl_sandbox_exec": "../../api/flytectl/gen/flytectl_sandbox_exec",

    # flytectl demo
    "flytectl/demo": "../../api/flytectl/demo.html",
    "flytectl/gen/flytectl_demo_start": "../../api/flytectl/gen/flytectl_demo_start",
    "flytectl/gen/flytectl_demo_status": "../../api/flytectl/gen/flytectl_demo_status",
    "flytectl/gen/flytectl_demo_teardown": "../../api/flytectl/gen/flytectl_demo_teardown",
    "flytectl/gen/flytectl_demo_exec": "../../api/flytectl/gen/flytectl_demo_exec",
    "flytectl/gen/flytectl_demo_reload": "../../api/flytectl/gen/flytectl_demo_reload",

    # flyteidl
    "reference_flyteidl": "../../api/flyteidl/docs_index.html",
    "protos/docs/core/core": "../../api/flyteidl/docs/core/core.html",
    "protos/docs/admin/admin": "../../api/flyteidl/docs/admin/admin.html",
    "protos/docs/service/service": "../../api/flyteidl/docs/service/service.html",
    "protos/docs/datacatalog/datacatalog": "../../api/flyteidl/docs/datacatalog/datacatalog.html",
    "protos/docs/event/event": "../../api/flyteidl/docs/event/event.html",
    "protos/docs/plugins/plugins": "../../api/flyteidl/docs/plugins/plugins.html",
    "protos/README": "../../api/flyteidl/contributing.html",
}


autosummary_generate = True
suppress_warnings = ["autosectionlabel.*", "myst.header"]
autodoc_typehints = "description"

# The master toctree document.
master_doc = "index"

# The language for content autogenerated by Sphinx. Refer to documentation
# for a list of supported languages.
#
# This is also used if you do content translation via gettext catalogs.
# Usually you set "language" from the command line for these cases.
language = "en"

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This pattern also affects html_static_path and html_extra_path .
exclude_patterns = [
    "_build",
    "Thumbs.db",
    ".DS_Store",
    "auto/**/*.ipynb",
    "auto/**/*.py",
    "auto/**/*.md",
    "examples/**/*.ipynb",
    "examples/**/*.py",
    "jupyter_execute/**",
    "README.md",
    "_projects/**",
    "_src/**",
    "examples/**",
    "flytesnacks/index.md",
    "flytesnacks/bioinformatics_examples.md",
    "flytesnacks/feature_engineering.md",
    "flytesnacks/flyte_lab.md",
    "flytesnacks/ml_training.md",
    "flytesnacks/deprecated_integrations.md",
    "flytesnacks/README.md",
    "flytekit/**/README.md",
    "flytekit/_templates/**",
    "examples/advanced_composition/**",
    "examples/basics/**",
    "examples/customizing_dependencies/**",
    "examples/data_types_and_io/**",
    "examples/development_lifecycle/**",
    "examples/extending/**",
    "examples/productionizing/**",
    "examples/testing/**",
    "flytesnacks/examples/advanced_composition/*.md",
    "flytesnacks/examples/basics/*.md",
    "flytesnacks/examples/customizing_dependencies/*.md",
    "flytesnacks/examples/data_types_and_io/*.md",
    "flytesnacks/examples/development_lifecycle/*.md",
    "flytesnacks/examples/extending/*.md",
    "flytesnacks/examples/productionizing/*.md",
    "flytesnacks/examples/testing/*.md",
    "api/flytectl/index.rst",
    "protos/boilerplate/**",
    "protos/tmp/**",
    "protos/gen/**",
    "protos/docs/**/index.rst",
    "protos/index.rst",
    "api/flytekit/_templates/**",
    "api/flytekit/index.rst",
]

# -- Options for HTML output -------------------------------------------------

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#
html_favicon = "images/favicon-flyte-docs.png"
html_logo = "images/favicon-flyte-docs.png"
html_theme = "pydata_sphinx_theme"
html_title = "Flyte"

templates_path = ["_templates"]

pygments_style = "tango"
pygments_dark_style = "native"

html_context = {
    "dir_to_title": {
        "api": "API reference",
        "deployment": "Deployment guide",
        "community": "Community",
        "concepts": "Flyte concepts",
        "ecosystem": "Ecosystem",
        "integrations": "Integrations",
        "tutorials": "Tutorials",
        "user_guide": "User guide",
    },
    "github_user": "flyteorg",
    "github_repo": "flyte",
    "github_version": "master",
    "doc_path": "docs",
}

html_theme_options = {
    # custom flyteorg pydata theme options
    # "github_url": "https://github.com/flyteorg/flyte",
    "logo": {
        "text": "Flyte",
    },
    "external_links": [
        {"name": "Flyte", "url": "https://flyte.org"},
    ],
    "icon_links": [
        {
            "name": "GitHub",
            "icon": "fa-brands fa-github",
            "type": "fontawesome",
            "url": "https://github.com/flyteorg/flyte",
        },
        {
            "name": "Slack",
            "url": "https://slack.flyte.org",
            "icon": "fa-brands fa-slack",
            "type": "fontawesome",
        },
        {
            "name": "Flyte",
            "url": "https://flyte.org",
            "icon": "fa-solid fa-dragon",
            "type": "fontawesome",
        }
    ],
    "use_edit_page_button": True,
    "navbar_start": ["navbar-logo"],
    "secondary_sidebar_items": ["page-toc", "edit-this-page"],
}

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
html_static_path = ["_static"]
html_css_files = ["custom.css"]
html_js_files = ["custom.js"]

# Custom sidebar templates, must be a dictionary that maps document names
# to template names.
#
# The default sidebars (for documents that don't match any pattern) are
# defined by theme itself.  Builtin themes are using these templates by
# default: ``['localtoc.html', 'relations.html', 'sourcelink.html',
# 'searchbox.html']``.
#
html_sidebars = {
    "api/**": ["sidebar/custom"],
    "deployment/**": ["sidebar/custom"],
    "community/**": ["sidebar/custom"],
    "concepts/**": ["sidebar/custom"],
    "ecosystem/**": ["sidebar/custom"],
    "flytesnacks/integrations/**": ["sidebar/integrations"],
    "flytesnacks/tutorials/**": ["sidebar/tutorials"],
    "user_guide/**": ["sidebar/custom"]
}

# -- Options for HTMLHelp output ---------------------------------------------

# Output file base name for HTML help builder.
htmlhelp_basename = "Flytedoc"

# -- Options for LaTeX output ------------------------------------------------

latex_elements = {
    # The paper size ('letterpaper' or 'a4paper').
    #
    # 'papersize': 'letterpaper',
    # The font size ('10pt', '11pt' or '12pt').
    #
    # 'pointsize': '10pt',
    # Additional stuff for the LaTeX preamble.
    #
    # 'preamble': '',
    # Latex figure (float) alignment
    #
    # 'figure_align': 'htbp',
}

# Grouping the document tree into LaTeX files. List of tuples
# (source start file, target name, title,
#  author, documentclass [howto, manual, or own class]).
latex_documents = [
    (master_doc, "Flyte.tex", "Flyte Documentation", "Flyte Authors", "manual"),
]

# -- Options for manual page output ------------------------------------------

# One entry per manual page. List of tuples
# (source start file, name, description, authors, manual section).
man_pages = [(master_doc, "flyte", "Flyte Documentation", [author], 1)]

# -- Options for Texinfo output ----------------------------------------------

# Grouping the document tree into Texinfo files. List of tuples
# (source start file, target name, title, author,
#  dir menu entry, description, category)
texinfo_documents = [
    (
        master_doc,
        "Flyte",
        "Flyte Documentation",
        author,
        "Flyte",
        "Accelerate your ML and data workflows to production.",
        "Miscellaneous",
    ),
]

# -- Extension configuration -------------------------------------------------
# autosectionlabel_prefix_document = True
autosectionlabel_maxdepth = 2

# Tags config
tags_create_tags = True
tags_extension = ["md", "rst"]
tags_page_title = "Tag"
tags_overview_title = "Pages by tags"

# Algolia Docsearch credentials
docsearch_app_id = "WLG0MZB58Q"
docsearch_api_key = os.getenv("DOCSEARCH_API_KEY")
docsearch_index_name = "flyte"

# -- Options for intersphinx extension ---------------------------------------

# Example configuration for intersphinx: refer to the Python standard library.
# intersphinx configuration. Uncomment the repeats with the local paths and update your username
# to help with local development.
intersphinx_mapping = {
    "python": ("https://docs.python.org/3", None),
    "numpy": ("https://numpy.org/doc/stable", None),
    "pandas": ("https://pandas.pydata.org/pandas-docs/stable/", None),
    "torch": ("https://pytorch.org/docs/main/", None),
    "scipy": ("https://docs.scipy.org/doc/scipy/reference", None),
    "matplotlib": ("https://matplotlib.org", None),
    "pandera": ("https://pandera.readthedocs.io/en/stable/", None),
}

myst_enable_extensions = ["colon_fence"]
myst_heading_anchors = 6

# Sphinx-mermaid config
mermaid_output_format = "raw"
mermaid_version = "latest"
mermaid_init_js = "mermaid.initialize({startOnLoad:true});"

# Makes it so that only the command is copied, not the output
copybutton_prompt_text = "$ "

# prevent css style tags from being copied by the copy button
copybutton_exclude = 'style[type="text/css"]'

nb_execution_mode = "off"
nb_execution_excludepatterns = [
    "flytekit/**/*",
    "flytesnacks/**/*",
    "examples/**/*",
]
nb_custom_formats = {
    ".md": ["jupytext.reads", {"fmt": "md:myst"}],
}

# Pattern for removing intersphinx references from source files.
# This should handle cases like:
#
# - :ref:`cookbook:label` -> :ref:`label`
# - :ref:`Text <cookbook:label>` -> :ref:`Text <label>`
INTERSPHINX_REFS_PATTERN = r"([`<])(flyte:|flytekit:|flytectl:|flyteidl:|cookbook:|idl:)"
INTERSPHINX_REFS_REPLACE = r"\1"

# Pattern for replacing all ref/doc labels that point to protos/docs with /protos/docs
PROTO_REF_PATTERN = r"([:<])(protos/docs)"
PROTO_REF_REPLACE = r"\1/api/flyteidl/docs"

# These patterns are used to replace values in source files that are imported
# from other repos.
REPLACE_PATTERNS = {

    r"<flyte:deployment/index>": r"</deployment/index>",
    r"<flytectl:index>": r"</api/flytectl/overview>",
    INTERSPHINX_REFS_PATTERN: INTERSPHINX_REFS_REPLACE,
    r"</auto_examples": r"<flytesnacks/examples",
    r"/auto_examples": r"/flytesnacks/examples",
    r"<protos/docs/core/core:taskmetadata>": r"<ref_flyteidl.core.TaskMetadata>",
    r"<protos/docs/core/core:tasktemplate>": r"<ref_flyteidl.core.TaskTemplate>",
    r"<flytesnacks/examples": r"</flytesnacks/examples",
    # r"</auto_examples/basics/index>": r"</flytesnacks/examples/basics/index>",
    r"<deploy-sandbox-local>": r"<deployment-deployment-sandbox>",
    r"<deployment/configuration/general:configurable resource types>": r"<deployment-configuration-general>",
    r"<_tags/DistributedComputing>": r"</_tags/DistributedComputing>",
    r"{ref}`bioinformatics <bioinformatics>`": r"bioinformatics",
    PROTO_REF_PATTERN: PROTO_REF_REPLACE,
    r"/protos/docs/service/index": r"/api/flyteidl/docs/service/service"
}

# r"<environment_setup>": r"</flytesnacks/environment_setup>",

import_projects_config = {
    "clone_dir": "_projects",
    "flytekit_api_dir": "_src/flytekit/",
    "source_regex_mapping": REPLACE_PATTERNS,
    "list_table_toc": [
       "flytesnacks/tutorials",
       "flytesnacks/integrations"
    ],
    "dev_build": bool(int(os.environ.get("MONODOCS_DEV_BUILD", 1))),
}

# Define these environment variables to use local copies of the projects. This
# is useful for building the docs in the CI/CD of the corresponding repos.
flytesnacks_local_path = os.environ.get("FLYTESNACKS_LOCAL_PATH", None)
flytekit_local_path = os.environ.get("FLYTEKIT_LOCAL_PATH", None)

flytesnacks_path = flytesnacks_local_path or "_projects/flytesnacks"
flytekit_path = flytekit_local_path or "_projects/api/flytekit"

import_projects = [
    {
        "name": "flytesnacks",
        "source": flytesnacks_local_path or "https://github.com/flyteorg/flytesnacks",
        "docs_path": "docs",
        "dest": "flytesnacks",
        "cmd": [
            ["cp", "-R", f"{flytesnacks_path}/examples", "./examples"],
            [
                # remove un-needed docs files in flytesnacks
                "rm",
                "-rf",
                "flytesnacks/jupyter_execute",
                "flytesnacks/auto_examples",
                "flytesnacks/_build",
                "flytesnacks/_tags",
                "flytesnacks/index.md",
            ]
        ],
        "local": flytesnacks_local_path is not None,
    },
    {
        "name": "mecoliflytesnacks",
        "source": "https://github.com/Mecoli1219/flytesnacks",
        "docs_path": "docs",
        "dest": "mecoliflytesnacks",
        "cmd": [
            ["git", "-C", f"{flytesnacks_path}", "checkout", "origin/jupyter-basic"],
            ["cp", "-R", f"{flytesnacks_path}/examples", "./mecoliexamples"],
            [
                # remove un-needed docs files in flytesnacks
                "rm",
                "-rf",
                "flytesnacks/jupyter_execute",
                "flytesnacks/auto_examples",
                "flytesnacks/_build",
                "flytesnacks/_tags",
                "flytesnacks/index.md",
            ]
        ],
        "local": False,
    },
    {
        "name": "flytekit",
        "source": flytekit_local_path or "https://github.com/flyteorg/flytekit",
        "docs_path": "docs/source",
        "dest": "api/flytekit",
        "cmd": [
            ["mkdir", "-p", import_projects_config["flytekit_api_dir"]],
            ["cp", "-R", f"{flytekit_path}/flytekit", import_projects_config["flytekit_api_dir"]],
            ["cp", "-R", f"{flytekit_path}/plugins", import_projects_config["flytekit_api_dir"]],
            ["cp", "-R", f"{flytekit_path}/tests", "./tests"],
        ],
        "local": flytekit_local_path is not None,
    },
    {
        "name": "flytectl",
        "source": "../flytectl",
        "docs_path": "docs/source",
        "dest": "api/flytectl",
        "local": True,
    },
    {
        "name": "flyteidl",
        "source": "../flyteidl",
        "docs_path": "protos",
        "dest": "api/flyteidl",
        "local": True,
    }
]

# myst notebook docs customization
auto_examples_dir_root = "examples"
auto_examples_dir_dest = "flytesnacks/examples"
auto_examples_refresh = False

# -- Options for todo extension ----------------------------------------------

# If true, `todo` and `todoList` produce output, else they produce nothing.
todo_include_todos = True

# -- Options for Sphinx issues extension --------------------------------------

# GitHub repo
issues_github_path = "flyteorg/flyte"

# equivalent to
issues_uri = "https://github.com/flyteorg/flyte/issues/{issue}"
issues_pr_uri = "https://github.com/flyteorg/flyte/pull/{pr}"
issues_commit_uri = "https://github.com/flyteorg/flyte/commit/{commit}"


# Disable warnings from flytekit
os.environ["FLYTE_SDK_LOGGING_LEVEL_ROOT"] = "50"

# Disable warnings from tensorflow
os.environ["TF_CPP_MIN_LOG_LEVEL"] = "3"

# Define the canonical URL if you are using a custom domain on Read the Docs
html_baseurl = os.environ.get("READTHEDOCS_CANONICAL_URL", "")

# Tell Jinja2 templates the build is running on Read the Docs
if os.environ.get("READTHEDOCS", "") == "True":
    if "html_context" not in globals():
        html_context = {}
    html_context["READTHEDOCS"] = True


class CustomWarningSuppressor(logging.Filter):
    """Filter logs by `suppress_warnings`."""

    def __init__(self, app: sphinx.application.Sphinx) -> None:
        self.app = app
        super().__init__()

    def filter(self, record: logging.LogRecord) -> bool:
        msg = record.getMessage()

        # TODO: These are all warnings that should be fixed as follow-ups to the
        # monodocs build project.
        filter_out = (
            "duplicate label",
            "Unexpected indentation",
            'Error with CSV data in "csv-table" directive',
            "Definition list ends without a blank line",
            "autodoc: failed to import module 'awssagemaker' from module 'flytekitplugins'",
            "Enumerated list ends without a blank line",
            'Unknown directive type "toc".',  # need to fix flytesnacks/contribute.md
        )

        if msg.strip().startswith(filter_out):
            return False

        if (
            msg.strip().startswith("document isn't included in any toctree")
            and record.location == "_tags/tagsindex"
        ):
            # ignore this warning, since we don't want the side nav to be
            # cluttered with the tags index page.
            return False

        return True


def linkcode_resolve(domain, info):
    """
    Determine the URL corresponding to Python object
    """
    if domain != "py":
        return None

    import flytekit

    modname = info["module"]
    fullname = info["fullname"]

    submod = sys.modules.get(modname)
    if submod is None:
        return None

    obj = submod
    for part in fullname.split("."):
        try:
            obj = getattr(obj, part)
        except AttributeError:
            return None

    if inspect.isfunction(obj):
        obj = inspect.unwrap(obj)
    try:
        fn = inspect.getsourcefile(obj)
    except TypeError:
        fn = None
    if not fn or fn.endswith("__init__.py"):
        try:
            fn = inspect.getsourcefile(sys.modules[obj.__module__])
        except (TypeError, AttributeError, KeyError):
            fn = None
    if not fn:
        return None

    try:
        source, lineno = inspect.getsourcelines(obj)
    except (OSError, TypeError):
        lineno = None

    linespec = f"#L{lineno:d}-L{lineno + len(source) - 1:d}" if lineno else ""

    startdir = Path(flytekit.__file__).parent.parent
    try:
        fn = os.path.relpath(fn, start=startdir).replace(os.path.sep, "/")
    except ValueError:
        return None

    if not fn.startswith("flytekit/"):
        return None

    if flytekit.__version__ == "dev":
        tag = "master"
    else:
        tag = f"v{flytekit.__version__}"

    return f"https://github.com/flyteorg/flytekit/blob/{tag}/{fn}{linespec}"


def setup(app: sphinx.application.Sphinx) -> None:
    """Setup root logger for Sphinx"""
    logger = logging.getLogger("sphinx")

    warning_handler, *_ = [
        h for h in logger.handlers
        if isinstance(h, sphinx_logging.WarningStreamHandler)
    ]
    warning_handler.filters.insert(0, CustomWarningSuppressor(app))

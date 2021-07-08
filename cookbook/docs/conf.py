# Configuration file for the Sphinx documentation builder.
#
# This file only contains a selection of the most common options. For a full
# list see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Path setup --------------------------------------------------------------

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
#
import logging
import os
import re
import sys
from pathlib import Path

from sphinx.errors import ConfigError
from sphinx_gallery.sorting import FileNameSortKey

sys.path.insert(0, os.path.abspath("../"))

# -- Project information -----------------------------------------------------

project = "Flyte Tutorial"
copyright = "2021, Flyte"
author = "Flyte"

# The full version, including alpha/beta/rc tags
# The full version, including alpha/beta/rc tags.
release = re.sub("^v", "", os.popen("git describe").read().strip())


class CustomSorter(FileNameSortKey):
    CUSTOM_FILE_SORT_ORDER = [
        # Flyte Basics
        "hello_world.py",
        "task.py",
        "basic_workflow.py",
        "imperative_wf_style.py",
        "lp.py",
        "task_cache.py",
        "files.py",
        "folders.py",
        "named_outputs.py"
        # Control Flow
        "run_conditions.py",
        "subworkflows.py",
        "dynamics.py",
        "map_task.py",
        "run_merge_sort.py",
        # Type System
        "flyte_python_types.py",
        "schema.py",
        "typed_schema.py",
        "custom_objects.py",
        "enums.py",
        # Testing
        "mocking.py",
        # Containerization
        "raw_container.py",
        "multi_images.py",
        "use_secrets.py",
        "spot_instances.py",
        "workflow_labels_annotations.py",
        # Deployment
        ## Workflow
        "deploying_workflows.py",
        "lp_schedules.py",
        "customizing_resources.py",
        "lp_notifications.py",
        "fast_registration.py",
        "multiple_k8s.py",
        ## Cluster
        "config_flyte_deploy.py",
        "productionize_cluster.py",
        "auth_setup.py",
        "auth_migration.py",
        "config_resource_mgr.py",
        "monitoring.py",
        "notifications.py",
        "optimize_perf.py",
        "access_cloud_resources.py",
        "auth_setup_appendix.py",
        ## Guides
        "kubernetes.py",
        "aws.py",
        "gcp.py",
        # Control Plane
        "register_project.py",
        "run_task.py",
        "run_workflow.py",
        # Integrations
        ## Flytekit Plugins
        "simple.py",
        "basic_schema_example.py",
        "branch_example.py",
        "quickstart_example.py",
        "dolt_quickstart_example.py",
        "dolt_branch_example.py",
        ## Kubernetes
        "pod.py",
        "pyspark_pi.py",
        "dataframe_passing.py",
        "pytorch_mnist.py",
        ## AWS
        "sagemaker_builtin_algo_training.py",
        "sagemaker_custom_training.py",
        "sagemaker_pytorch_distributed_training.py",
        ## GCP
        # TODO
        ## External Services
        "hive.py"
        # Extending Flyte
        "backend_plugins.py",  # NOTE: for some reason this needs to be listed first here to show up last on the TOC
        "run_custom_types.py",
        "custom_task_plugin.py",
        # Tutorials
        ## ML Training
        "diabetes.py",
        "house_price_predictor.py",
        "multiregion_house_price_predictor.py",
        "datacleaning_tasks.py",
        "datacleaning_workflow.py",
    ]
    """
    Take a look at the code for the default sorter included in the sphinx_gallery to see how this works.
    """

    def __call__(self, filename):
        src_file = os.path.normpath(os.path.join(self.src_dir, filename))
        if filename in self.CUSTOM_FILE_SORT_ORDER:
            return f"{self.CUSTOM_FILE_SORT_ORDER.index(filename):03d}"
        else:
            logging.warning(
                f"File {filename} not found in static ordering list, temporarily adding to the end"
            )
            self.CUSTOM_FILE_SORT_ORDER.append(src_file)
            return f"{len(self.CUSTOM_FILE_SORT_ORDER) - 1:03d}"


# -- General configuration ---------------------------------------------------

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
extensions = [
    "sphinx.ext.autodoc",
    "sphinx.ext.autosummary",
    "sphinx.ext.autosectionlabel",
    "sphinx.ext.napoleon",
    "sphinx.ext.todo",
    "sphinx.ext.viewcode",
    "sphinx.ext.doctest",
    "sphinx.ext.intersphinx",
    "sphinx.ext.coverage",
    "sphinx_gallery.gen_gallery",
    "sphinx-prompt",
    "sphinx_copybutton",
    "sphinxext.remoteliteralinclude",
    "sphinx_panels",
    "sphinx_tabs.tabs",
    "sphinxcontrib.mermaid",
]

# Add any paths that contain templates here, relative to this directory.
templates_path = ["_templates"]

html_static_path = ["_static"]

html_css_files = ["sphx_gallery_autogen.css"]

# generate autosummary even if no references
autosummary_generate = True

# The suffix of source filenames.
source_suffix = ".rst"

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This pattern also affects html_static_path and html_extra_path.
exclude_patterns = ["_build", "Thumbs.db", ".DS_Store"]

# The master toctree document.
master_doc = "index"

pygments_style = "tango"
pygments_dark_style = "monokai"

html_context = {
    "home_page": "https://docs.flyte.org",
}

# -- Options for HTML output -------------------------------------------------

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#
html_theme = "furo"
html_title = "Flyte Docs"

html_theme_options = {
    "light_css_variables": {
        "color-brand-primary": "#4300c9",
        "color-brand-content": "#4300c9",
    },
    "dark_css_variables": {
        "color-brand-primary": "#9D68E4",
        "color-brand-content": "#9D68E4",
    },
    # custom flyteorg furo theme options
    "github_repo": "flytesnacks",
    "github_username": "flyteorg",
    "github_commit": "master",
    "docs_path": "cookbook/docs",  # path to documentation source
    "sphinx_gallery_src_dir": "cookbook",  # path to directory of sphinx gallery source files relative to repo root
    "sphinx_gallery_dest_dir": "auto",  # path to root directory containing auto-generated sphinx gallery examples
}

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".

html_favicon = "_static/flyte_circle_gradient_1_4x4.png"
html_logo = "_static/flyte_circle_gradient_1_4x4.png"

examples_dirs = [
    "../core/flyte_basics",
    "../core/control_flow",
    "../core/type_system",
    "../case_studies/ml_training/pima_diabetes",
    "../case_studies/ml_training/house_price_prediction",
    "../case_studies/feature_engineering/sqlite_datacleaning",
    "../testing",
    "../core/containerization",
    "../deployment",
    # "../control_plane",  # TODO: add content to this section
    # "../integrations/flytekit_plugins/sqllite3",  # TODO: add content to this section
    "../integrations/flytekit_plugins/papermilltasks",
    # "../integrations/flytekit_plugins/sqlalchemy",  # TODO: add content to this section
    "../integrations/flytekit_plugins/pandera",
    "../integrations/flytekit_plugins/dolt",
    "../integrations/kubernetes/pod",
    "../integrations/kubernetes/k8s_spark",
    # "../integrations/kubernetes/kftensorflow",  # TODO: need to update content
    "../integrations/kubernetes/kfpytorch",
    # "../integrations/aws/athena",  # TODO: add content to this section
    "../integrations/aws/sagemaker_training",
    "../integrations/aws/sagemaker_pytorch",
    "../integrations/gcp",
    "../integrations/external_services/hive",
    "../core/extend_flyte",
]
gallery_dirs = [
    "auto/core/flyte_basics",
    "auto/core/control_flow",
    "auto/core/type_system",
    "auto/case_studies/ml_training/pima_diabetes",
    "auto/case_studies/ml_training/house_price_prediction",
    "auto/case_studies/feature_engineering/sqlite_datacleaning",
    "auto/testing",
    "auto/core/containerization",
    "auto/deployment",
    # "auto/deployment/guides",  # TODO: add content to this section
    # "auto/control_plane",  # TODO: add content to this section
    # "auto/integrations/flytekit_plugins/sqllite3",  # TODO: add content to this section
    "auto/integrations/flytekit_plugins/papermilltasks",
    # "auto/integrations/flytekit_plugins/sqlalchemy",  # TODO: add content to this section
    "auto/integrations/flytekit_plugins/pandera",
    "auto/integrations/flytekit_plugins/dolt",
    "auto/integrations/kubernetes/pod",
    "auto/integrations/kubernetes/k8s_spark",
    # "auto/integrations/kubernetes/kftensorflow",  # TODO: need to update content
    "auto/integrations/kubernetes/kfpytorch",
    # "auto/integrations/aws/athena",  # TODO: add content to this section
    "auto/integrations/aws/sagemaker_training",
    "auto/integrations/aws/sagemaker_pytorch",
    "auto/integrations/gcp",
    "auto/integrations/external_services/hive",
    "auto/core/extend_flyte",
]

# image_scrapers = ('matplotlib',)
image_scrapers = ()

min_reported_time = 0

# hide example pages with empty content
ignore_py_files = [
    "__init__",
    "config_resource_mgr",
    "optimize_perf",
]

sphinx_gallery_conf = {
    "examples_dirs": examples_dirs,
    "gallery_dirs": gallery_dirs,
    "ignore_pattern": f"({'|'.join(ignore_py_files)})\.py",
    # "subsection_order": ExplicitOrder(
    #     [
    #         "../core/basic",
    #         "../core/intermediate",
    #         "../core/advanced",
    #         "../core/remote_flyte",
    #         "../case_studies/pima_diabetes",
    #         "../case_studies/house_price_prediction",
    #         "../testing",
    #         "../plugins/pod/",
    #         "../plugins/k8s_spark",
    #         "../plugins/papermilltasks/",
    #         "../plugins/hive",
    #         "../plugins/sagemaker_training",
    #         "../plugins/kfpytorch",
    #         "../plugins/sagemaker_pytorch/",
    #     ]
    # ),
    # # specify the order of examples to be according to filename
    "within_subsection_order": CustomSorter,
    "min_reported_time": min_reported_time,
    "filename_pattern": "/run_",
    "capture_repr": (),
    "image_scrapers": image_scrapers,
    "default_thumb_file": "flyte_mark_offset_pink.png",
    "thumbnail_size": (350, 350),
    # Support for binder
    # 'binder': {'org': 'sphinx-gallery',
    # 'repo': 'sphinx-gallery.github.io',
    # 'branch': 'master',
    # 'binderhub_url': 'https://mybinder.org',
    # 'dependencies': './binder/requirements.txt',
    # 'notebooks_dir': 'notebooks',
    # 'use_jupyter_lab': True,
    # },
}

if len(examples_dirs) != len(gallery_dirs):
    raise ConfigError("examples_dirs and gallery_dirs aren't of the same length")

# Sphinx gallery makes specific assumptions about the structure of example gallery.
# The main one is the the gallery's entrypoint is a README.rst file and the rest
# of the files are *.py files that are auto-converted to .rst files. This makes
# sure that the only rst files in the example directories are README.rst
hide_download_page_ids = []


def hide_example_page(file_handler):
    """Heuristic that determines whether example file contains python code."""
    example_content = file_handler.read().strip()

    no_percent_comments = True
    no_imports = True

    for line in example_content.split("\n"):
        if line.startswith(r"# %%"):
            no_percent_comments = False
        if line.startswith("import"):
            no_imports = False

    return (
            example_content.startswith('"""')
            and example_content.endswith('"""')
            and no_percent_comments
            and no_imports
    )


for source_dir in sphinx_gallery_conf["examples_dirs"]:
    for f in Path(source_dir).glob("*.rst"):
        if f.name != "README.rst":
            raise ValueError(
                f"non-README.rst file {f} not permitted in sphinx gallery directories"
            )

    # we want to hide the download example button in pages that don't actually contain python code.
    for f in Path(source_dir).glob("*.py"):
        with f.open() as fh:
            if hide_example_page(fh):
                page_id = (
                    str(f)
                        .replace("..", "auto")
                        .replace("/", "-")
                        .replace(".", "-")
                        .replace("_", "-")
                )
                hide_download_page_ids.append(f"sphx-glr-download-{page_id}")

SPHX_GALLERY_CSS_TEMPLATE = """
{hide_download_page_ids} {{
    height: 0px;
    visibility: hidden;
}}
"""

with Path("_static/sphx_gallery_autogen.css").open("w") as f:
    f.write(
        SPHX_GALLERY_CSS_TEMPLATE.format(
            hide_download_page_ids=",\n".join(f"#{x}" for x in hide_download_page_ids)
        )
    )

# intersphinx configuration
intersphinx_mapping = {
    "python": ("https://docs.python.org/{.major}".format(sys.version_info), None),
    "numpy": ("https://numpy.org/doc/stable", None),
    "pandas": ("https://pandas.pydata.org/pandas-docs/stable/", None),
    "pandera": ("https://pandera.readthedocs.io/en/stable/", None),
    "dolt": ("https://docs.dolthub.com/", None),
    "torch": ("https://pytorch.org/docs/master/", None),
    "scipy": ("https://docs.scipy.org/doc/scipy/reference", None),
    "matplotlib": ("https://matplotlib.org", None),
    "flytekit": ("https://flyte.readthedocs.io/projects/flytekit/en/latest/", None),
    "flyte": ("https://flyte.readthedocs.io/en/latest/", None),
    # Uncomment for local development and change to your username
    # "flytekit": ("/Users/ytong/go/src/github.com/lyft/flytekit/docs/build/html", None),
    "flyteidl": ("https://docs.flyte.org/projects/flyteidl/en/latest", None),
    "flytectl": ("https://docs.flyte.org/projects/flytectl/en/latest/", None),
}

# Sphinx-tabs config
sphinx_tabs_valid_builders = ['linkcheck']

# Sphinx-mermaid config
mermaid_output_format = 'raw'
mermaid_version = 'latest'
mermaid_init_js = "mermaid.initialize({startOnLoad:false});"

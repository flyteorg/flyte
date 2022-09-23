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
sys.path.append(os.path.abspath("./_ext"))

# -- Project information -----------------------------------------------------

project = "Flytesnacks"
copyright = "2022, Flyte"
author = "Flyte"

# The full version, including alpha/beta/rc tags
release = re.sub("^v", "", os.popen("git describe").read().strip())


class CustomSorter(FileNameSortKey):
    CUSTOM_FILE_SORT_ORDER = [
        # Flyte Basics
        "hello_world.py",
        "task.py",
        "basic_workflow.py",
        "imperative_wf_style.py",
        "documented_workflow.py",
        "lp.py",
        "deck.py",
        "task_cache.py",
        "deck.py",
        "task_cache.py",
        "shell_task.py",
        "reference_task.py",
        "files.py",
        "folders.py",
        "named_outputs.py",
        "decorating_tasks.py",
        "decorating_workflows.py",
        # Control Flow
        "conditions.py",
        "chain_entities.py",
        "subworkflows.py",
        "dynamics.py",
        "map_task.py",
        "checkpoint.py",
        "merge_sort.py",
        # Type System
        "flyte_python_types.py",
        "schema.py",
        "structured_dataset.py",
        "typed_schema.py",
        "pytorch_types.py",
        "custom_objects.py",
        "enums.py",
        "lp_schedules.py",
        # Testing
        "mocking.py",
        # Containerization
        "raw_container.py",
        "multi_images.py",
        "use_secrets.py",
        "spot_instances.py",
        "workflow_labels_annotations.py",
        # Remote Access
        "register_project.py",
        "remote_task.py",
        "remote_workflow.py",
        "remote_launchplan.py",
        "inspecting_executions.py",
        "debugging_workflows_tasks.py",
        # Deployment
        ## Workflow
        "deploying_workflows.py",
        "customizing_resources.py",
        "lp_notifications.py",
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
        # Integrations
        ## Flytekit Plugins
        "simple.py",
        "basic_schema_example.py",
        "branch_example.py",
        "quickstart_example.py",
        "dolt_quickstart_example.py",
        "dolt_branch_example.py",
        "task_example.py",
        "type_example.py",
        "knn_classifier.py",
        "sqlite3_integration.py",
        "sql_alchemy.py",
        "whylogs_example.py",
        ## Kubernetes
        "pod.py",
        "pyspark_pi.py",
        "dataframe_passing.py",
        "pytorch_mnist.py",
        "tf_mnist.py",
        ## AWS
        "sagemaker_builtin_algo_training.py",
        "sagemaker_custom_training.py",
        "sagemaker_pytorch_distributed_training.py",
        ## GCP
        "bigquery.py",
        ## External Services
        "hive.py",
        "snowflake.py",
        "airflow.py",
        # Extending Flyte
        "backend_plugins.py",  # NOTE: for some reason this needs to be listed first here to show up last on the TOC
        "custom_types.py",
        "custom_task_plugin.py",
        # Repo-based Projects
        "larger_apps_setup.py",
        "larger_apps_deploy.py",
        "larger_apps_iterate.py",
        # Tutorials
        ## ML Training
        "diabetes.py",
        "house_price_predictor.py",
        "multiregion_house_price_predictor.py",
        "keras_spark_rossmann_estimator.py",
        ## Feature Engineering
        "pytorch_single_node_and_gpu.py",
        "pytorch_single_node_multi_gpu.py",
        "notebook.py",
        "notebook_and_task.py",
        "notebook_as_tasks.py",
        "feature_eng_tasks.py",
        "feast_workflow.py",
        ## Bioinformatics
        "blastx_example.py",
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
    "sphinxcontrib.mermaid",
    "sphinxcontrib.yt",
    "sphinx_tabs.tabs",
    "run_example_cmds_extension",
]

# Add any paths that contain templates here, relative to this directory.
templates_path = ["_templates"]

html_static_path = ["_static"]
html_css_files = ["sphx_gallery_autogen.css", "custom.css"]

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
html_title = "Flyte"

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
    "../core/scheduled_workflows",
    "../core/type_system",
    "../case_studies/ml_training/pima_diabetes",
    "../case_studies/ml_training/house_price_prediction",
    "../case_studies/ml_training/mnist_classifier",
    "../case_studies/ml_training/spark_horovod",
    "../case_studies/feature_engineering/eda",
    "../case_studies/feature_engineering/feast_integration",
    "../case_studies/bioinformatics/blast",
    "../testing",
    "../core/containerization",
    "../deployment",
    "../remote_access",
    "../integrations/flytekit_plugins/sql",
    "../integrations/flytekit_plugins/greatexpectations",
    "../integrations/flytekit_plugins/papermilltasks",
    "../integrations/flytekit_plugins/pandera_examples",
    "../integrations/flytekit_plugins/modin_examples",
    "../integrations/flytekit_plugins/dolt",
    "../integrations/flytekit_plugins/whylogs_examples",
    "../integrations/flytekit_plugins/onnx_examples",
    "../integrations/kubernetes/pod",
    "../integrations/kubernetes/k8s_spark",
    "../integrations/kubernetes/kftensorflow",
    "../integrations/kubernetes/kfpytorch",
    "../integrations/kubernetes/kfmpi",
    "../integrations/kubernetes/ray_example",
    "../integrations/aws/athena",
    "../integrations/aws/batch",
    "../integrations/aws/sagemaker_training",
    "../integrations/aws/sagemaker_pytorch",
    "../integrations/gcp/bigquery",
    "../integrations/external_services/hive",
    "../integrations/external_services/snowflake",
    "../integrations/external_services/airflow",
    "../core/extend_flyte",
    "../larger_apps",
]
gallery_dirs = [
    "auto/core/flyte_basics",
    "auto/core/control_flow",
    "auto/core/scheduled_workflows",
    "auto/core/type_system",
    "auto/case_studies/ml_training/pima_diabetes",
    "auto/case_studies/ml_training/house_price_prediction",
    "auto/case_studies/ml_training/mnist_classifier",
    "auto/case_studies/ml_training/spark_horovod",
    "auto/case_studies/feature_engineering/eda",
    "auto/case_studies/feature_engineering/feast_integration",
    "auto/case_studies/bioinformatics/blast",
    "auto/testing",
    "auto/core/containerization",
    "auto/deployment",
    "auto/remote_access",
    "auto/integrations/flytekit_plugins/sql",
    "auto/integrations/flytekit_plugins/greatexpectations",
    "auto/integrations/flytekit_plugins/papermilltasks",
    "auto/integrations/flytekit_plugins/pandera_examples",
    "auto/integrations/flytekit_plugins/modin_examples",
    "auto/integrations/flytekit_plugins/dolt",
    "auto/integrations/flytekit_plugins/whylogs_examples",
    "auto/integrations/flytekit_plugins/onnx_examples",
    "auto/integrations/kubernetes/pod",
    "auto/integrations/kubernetes/k8s_spark",
    "auto/integrations/kubernetes/kftensorflow",
    "auto/integrations/kubernetes/kfpytorch",
    "auto/integrations/kubernetes/kfmpi",
    "auto/integrations/kubernetes/ray_example",
    "auto/integrations/aws/athena",
    "auto/integrations/aws/batch",
    "auto/integrations/aws/sagemaker_training",
    "auto/integrations/aws/sagemaker_pytorch",
    "auto/integrations/gcp/bigquery",
    "auto/integrations/external_services/hive",
    "auto/integrations/external_services/snowflake",
    "auto/integrations/external_services/airflow",
    "auto/core/extend_flyte",
    "auto/larger_apps",
]

# image_scrapers = ('matplotlib',)
image_scrapers = ()

min_reported_time = 0

# hide example pages with empty content
ignore_py_files = [
    r"__init__\.py",
    r"config_resource_mgr\.py",
    r"optimize_perf\.py",
    r"^run_.+\.py",
]

sphinx_gallery_conf = {
    "examples_dirs": examples_dirs,
    "gallery_dirs": gallery_dirs,
    "ignore_pattern": f"{'|'.join(ignore_py_files)}",
    # specify the order of examples to be according to filename
    "within_subsection_order": CustomSorter,
    "min_reported_time": min_reported_time,
    "capture_repr": (),
    "image_scrapers": image_scrapers,
    "default_thumb_file": "_static/code-example-icon.png",
    "thumbnail_size": (350, 350),
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
    "modin": ("https://modin.readthedocs.io/en/stable/", None),
    "torch": ("https://pytorch.org/docs/master/", None),
    "scipy": ("https://docs.scipy.org/doc/scipy/reference", None),
    "matplotlib": ("https://matplotlib.org", None),
    "flytekit": ("https://flyte.readthedocs.io/projects/flytekit/en/latest/", None),
    "flytekitplugins": ("https://docs.flyte.org/projects/flytekit/en/latest/", None),
    "flyte": ("https://flyte.readthedocs.io/en/latest/", None),
    # Uncomment for local development and change to your username
    # "flytekit": ("/Users/ytong/go/src/github.com/lyft/flytekit/docs/build/html", None),
    "flyteidl": ("https://docs.flyte.org/projects/flyteidl/en/latest", None),
    "flytectl": ("https://docs.flyte.org/projects/flytectl/en/latest/", None),
    "pytorch": ("https://pytorch.org/docs/stable/", None),
    "greatexpectations": ("https://legacy.docs.greatexpectations.io/en/latest", None),
    "tensorflow": (
        "https://www.tensorflow.org/api_docs/python",
        "https://github.com/GPflow/tensorflow-intersphinx/raw/master/tf2_py_objects.inv",
    ),
    "whylogs": ("https://whylogs.readthedocs.io/", None),
}

# Sphinx-tabs config
# sphinx_tabs_valid_builders = ["linkcheck"]

# Sphinx-mermaid config
mermaid_output_format = "raw"
mermaid_version = "latest"
mermaid_init_js = "mermaid.initialize({startOnLoad:false});"

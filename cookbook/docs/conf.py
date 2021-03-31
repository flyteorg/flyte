# Configuration file for the Sphinx documentation builder.
#
# This file only contains a selection of the most common options. For a full
# list see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Path setup --------------------------------------------------------------

import logging

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
#
import os
import re
import sys

from sphinx_gallery.sorting import ExplicitOrder, FileNameSortKey

sys.path.insert(0, os.path.abspath("../"))

# -- Project information -----------------------------------------------------

project = "Flyte Tutorial"
copyright = "2021, Flyte"
author = "Flyte"

# The full version, including alpha/beta/rc tags
# The full version, including alpha/beta/rc tags.
release = re.sub("^v", "", os.popen("git describe").read().strip())


class CustomSorter(FileNameSortKey):
    """
    Take a look at the code for the default sorter included in the sphinx_gallery to see how this works.
    """

    CUSTOM_FILE_SORT_ORDER = [
        # Basic
        "task.py",
        "hello_world.py",
        "basic_workflow.py",
        "lp.py",
        "task_cache.py",
        "files.py",
        "folders.py",
        "mocking.py",
        "diabetes.py",
        # Intermediate
        "schema.py",
        "typed_schema.py",
        "subworkflows.py",
        "dataframe_passing.py",
        "dynamics.py",
        "hive.py",
        "pyspark_pi.py",
        "custom_objects.py",
        "run_conditions.py",
        "raw_container.py",
        "map_task.py"
        # Advanced
        "run_merge_sort.py",
        "custom_task_plugin.py",
        "run_custom_types.py",
        # Remote Flyte
        "multi_images.py",
        "customizing_resources.py",
        "lp_schedules.py",
        "lp_notifications.py",
        # Native Plugins
        "pytorch_mnist.py",
        # AWS Plugins
        "sagemaker_builtin_algo_training.py",
        "sagemaker_custom_training.py",
    ]

    def __call__(self, filename):
        src_file = os.path.normpath(os.path.join(self.src_dir, filename))
        if filename in self.CUSTOM_FILE_SORT_ORDER:
            return f"{self.CUSTOM_FILE_SORT_ORDER.index(filename):03d}"
        else:
            logging.warning(
                f"File {filename} not found in static ordering list, temporarily adding to the end"
            )
            self.CUSTOM_FILE_SORT_ORDER.append(src_file)
            return f"{len(self.CUSTOM_FILE_SORT_ORDER)-1:03d}"


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
    "sphinx_search.extension",
]

# Add any paths that contain templates here, relative to this directory.
templates_path = ["_templates"]

html_static_path = ["_static"]

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
pygments_dark_style = "native"

html_css_files = ["custom.css"]

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
}

# html_sidebars = {
#     "**": ["logo-text.html", "globaltoc.html", "localtoc.html", "searchbox.html"]
# }

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".

html_logo = "flyte_circle_gradient_1_4x4.png"

examples_dirs = [
    "../core/basic",
    "../core/intermediate",
    "../core/advanced",
    "../core/remote_flyte",
    "../case_studies/pima_diabetes",
    "../plugins/hive",
    "../plugins/sagemaker_training",
    "../plugins/k8s_spark",
    "../plugins/kfpytorch",
    "../plugins/pod/",
    "../plugins/papermilltasks/",
    "../plugins/sagemaker_pytorch/",
]
gallery_dirs = [
    "auto_core_basic",
    "auto_core_intermediate",
    "auto_core_advanced",
    "auto_core_remote_flyte",
    "auto_case_studies",
    "auto_plugins_hive",
    "auto_plugins_sagemaker_training",
    "auto_plugins_k8s_spark",
    "auto_plugins_kfpytorch",
    "auto_plugins_pod",
    "auto_plugins_papermilltasks",
    "auto_plugins_sagemaker_pytorch",
]

# image_scrapers = ('matplotlib',)
image_scrapers = ()

min_reported_time = 0

sphinx_gallery_conf = {
    "examples_dirs": examples_dirs,
    "gallery_dirs": gallery_dirs,
    "subsection_order": ExplicitOrder(
        [
            "../core/basic",
            "../core/intermediate",
            "../core/advanced",
            "../core/remote_flyte",
            "../case_studies/pima_diabetes",
            "../plugins/pod/",
            "../plugins/k8s_spark",
            "../plugins/papermilltasks/",
            "../plugins/hive",
            "../plugins/sagemaker_training",
            "../plugins/kfpytorch",
            "../plugins/sagemaker_pytorch/",
        ]
    ),
    # specify the order of examples to be according to filename
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

# intersphinx configuration
intersphinx_mapping = {
    "python": ("https://docs.python.org/{.major}".format(sys.version_info), None),
    "numpy": ("https://numpy.org/doc/stable", None),
    "pandas": ("https://pandas.pydata.org/pandas-docs/stable/", None),
    "torch": ("https://pytorch.org/docs/master/", None),
    "scipy": ("https://docs.scipy.org/doc/scipy/reference", None),
    "matplotlib": ("https://matplotlib.org", None),
    "flytekit": ("https://flyte.readthedocs.io/projects/flytekit/en/master/", None),
    "flyte": ("https://flyte.readthedocs.io/en/latest/", None),
    # Uncomment for local development and change to your username
    # "flytekit": ("/Users/ytong/go/src/github.com/lyft/flytekit/docs/build/html", None),
}

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
import os
import sys
from sphinx_gallery.sorting import ExampleTitleSortKey

sys.path.insert(0, os.path.abspath("../"))

# -- Project information -----------------------------------------------------

project = "Flyte Cookbook 2nd Ed."
copyright = "2020, Flyte"
author = "Flyte"

# The full version, including alpha/beta/rc tags
release = "0.16.0"

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
]

# Add any paths that contain templates here, relative to this directory.
templates_path = ["_templates"]

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

# -- Options for HTML output -------------------------------------------------

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#
html_theme = "sphinx_rtd_theme"

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
html_static_path = ["_static"]

html_logo = "flyte_lockup_on_dark.png"
html_theme_options = {
    "logo_only": True,
    "display_version": True,
}


examples_dirs = [
    "../recipes",
]
gallery_dirs = ["auto_recipes"]

# image_scrapers = ('matplotlib',)
image_scrapers = ()

min_reported_time = 0

sphinx_gallery_conf = {
    "examples_dirs": examples_dirs,
    "gallery_dirs": gallery_dirs,
    # specify the order of examples to be according to filename
    "within_subsection_order": ExampleTitleSortKey,
    "min_reported_time": min_reported_time,
    "filename_pattern": "/run_",
    "capture_repr": (),
    "image_scrapers": image_scrapers,
    "default_thumb_file": "flyte_lockup_on_dark.png",
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
}

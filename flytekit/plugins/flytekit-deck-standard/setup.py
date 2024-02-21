from setuptools import setup

PLUGIN_NAME = "deck"

microlib_name = f"flytekitplugins-{PLUGIN_NAME}-standard"

plugin_requires = [
    "flytekit",
    "markdown",
    "plotly",
    # ydata-profiling is not compatible with python 3.12 yet: https://github.com/ydataai/ydata-profiling/issues/1510
    "ydata-profiling; python_version<'3.12'",
    "pandas",
    "ipywidgets",
    "pygments",
]

__version__ = "0.0.0+develop"

setup(
    name=microlib_name,
    version=__version__,
    author="flyteorg",
    author_email="admin@flyte.org",
    description="This Plugin provides more renderers to improve task visibility",
    url="https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-data-fsspec",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    namespace_packages=["flytekitplugins"],
    packages=[f"flytekitplugins.{PLUGIN_NAME}"],
    install_requires=plugin_requires,
    license="apache2",
    python_requires=">=3.8",
    classifiers=[
        "Intended Audience :: Science/Research",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: Apache Software License",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Topic :: Scientific/Engineering",
        "Topic :: Scientific/Engineering :: Artificial Intelligence",
        "Topic :: Software Development",
        "Topic :: Software Development :: Libraries",
        "Topic :: Software Development :: Libraries :: Python Modules",
    ],
    entry_points={"flytekit.plugins": [f"{PLUGIN_NAME}=flytekitplugins.{PLUGIN_NAME}"]},
)

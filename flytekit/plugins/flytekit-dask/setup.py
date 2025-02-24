from setuptools import setup

PLUGIN_NAME = "dask"

microlib_name = f"flytekitplugins-{PLUGIN_NAME}"

plugin_requires = [
    "flyteidl>=1.3.2",
    "flytekit>=1.3.0b2,<2.0.0",
    "dask[distributed]>=2022.10.2",
]

__version__ = "0.0.0+develop"

setup(
    name=microlib_name,
    version=__version__,
    author="flyteorg",
    author_email="admin@flyte.org",
    description="Dask plugin for flytekit",
    url="https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-dask",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    namespace_packages=["flytekitplugins"],
    packages=[f"flytekitplugins.{PLUGIN_NAME}"],
    install_requires=plugin_requires,
    license="apache2",
    python_requires=">=3.8",  # dask requires >= 3.8
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
)

from setuptools import setup

PLUGIN_NAME = "fsspec"

microlib_name = f"flytekitplugins-data-{PLUGIN_NAME}"

plugin_requires = []

__version__ = "0.0.0+develop"

setup(
    name=microlib_name,
    version=__version__,
    author="flyteorg",
    author_email="admin@flyte.org",
    description="This is a deprecated plugin as of flytekit 1.5",
    url="https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-data-fsspec",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    namespace_packages=["flytekitplugins"],
    packages=[f"flytekitplugins.{PLUGIN_NAME}"],
    install_requires=plugin_requires,
    extras_require={
        # https://github.com/fsspec/filesystem_spec/blob/master/setup.py#L36
        "abfs": [],
        "aws": [],
        "gcp": [],
    },
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
)

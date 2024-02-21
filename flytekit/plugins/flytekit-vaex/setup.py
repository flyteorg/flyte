from setuptools import setup

PLUGIN_NAME = "vaex"

microlib_name = f"flytekitplugins-{PLUGIN_NAME}"

# vaex doesn't support pydantic 2.0 yet. https://github.com/vaexio/vaex/issues/2384
plugin_requires = [
    "flytekit>=1.3.0b2,<2.0.0",
    "vaex-core>=4.13.0,<4.14; python_version < '3.10'",
    "vaex-core>=4.16.0; python_version >= '3.10'",
    "pandas",
    "pydantic<2.0",
]

__version__ = "0.0.0+develop"

setup(
    name=microlib_name,
    version=__version__,
    author="admin@flyte.org",
    description="Vaex plugin for flytekit",
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

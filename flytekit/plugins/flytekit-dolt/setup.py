import pip
from setuptools import setup
from setuptools.command.develop import develop

PLUGIN_NAME = "dolt"

microlib_name = f"flytekitplugins-{PLUGIN_NAME}"

plugin_requires = ["flytekit>=1.3.0b2,<2.0.0", "dolt_integrations>=0.1.5", "networkx<3.2; python_version<'3.9'"]
dev_requires = ["pytest-mock>=3.6.1"]

__version__ = "0.0.0+develop"


class PostDevelopCommand(develop):
    """Post-installation for development mode."""

    def run(self):
        develop.run(self)
        pip.main(["install"] + dev_requires)


setup(
    name=microlib_name,
    version=__version__,
    author="dolthub",
    author_email="max@dolthub.com",
    description="Dolt plugin for flytekit",
    namespace_packages=["flytekitplugins"],
    packages=[f"flytekitplugins.{PLUGIN_NAME}"],
    install_requires=plugin_requires,
    extras_resquire=dict(
        dev=dev_requires,
    ),
    cmdclass=dict(develop=PostDevelopCommand),
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
    scripts=["scripts/flytekit_install_dolt.sh"],
)

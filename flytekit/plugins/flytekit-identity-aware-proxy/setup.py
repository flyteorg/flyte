from setuptools import setup

PLUGIN_NAME = "identity_aware_proxy"

microlib_name = f"flytekitplugins-{PLUGIN_NAME}"

plugin_requires = [
    "click",
    "google-cloud-secret-manager",
    "google-auth",
    "flytekit>=1.10",
    # https://github.com/grpc/grpc/issues/33935
    # https://github.com/grpc/grpc/issues/35323
    "grpcio<1.55.0",
]

__version__ = "0.0.0+develop"

setup(
    name=microlib_name,
    version=__version__,
    author="flyteorg",
    author_email="admin@flyte.org",
    description="External command plugin to generate ID tokens for GCP Identity Aware Proxy",
    url="https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-identity-aware-proxy",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    namespace_packages=["flytekitplugins"],
    packages=[f"flytekitplugins.{PLUGIN_NAME}"],
    entry_points={
        "console_scripts": [
            "flyte-iap=flytekitplugins.identity_aware_proxy.cli:cli",
        ],
    },
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
)

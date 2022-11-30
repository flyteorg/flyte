from setuptools import setup, find_packages

__version__ = "0.0.0+develop"

setup(
    name='flyteidl',
    version=__version__,
    description='IDL for Flyte Platform',
    url='https://www.github.com/flyteorg/flyteidl',
    maintainer='FlyteOrg',
    maintainer_email='admin@flyte.org',
    packages=find_packages('gen/pb_python'),
    package_dir={'': 'gen/pb_python'},
    # https://github.com/pypa/setuptools/issues/3136 describes an extension to
    # setuptools that would involve a simpler way to specify this, but while
    # that does not happen we have to package the pyi files manually like so:
    package_data={'flyteidl': ["*.pyi", "**/*.pyi"]},
    dependency_links=[],
    install_requires=[
        'googleapis-common-protos',
        'protoc_gen_swagger',
        'protobuf>=4.21.1,<5.0.0',
        # Packages in here should rarely be pinned. This is because these
        # packages (at the specified version) are required for project
        # consuming this library. By pinning to a specific version you are the
        # number of projects that can consume this or forcing them to
        # upgrade/downgrade any dependencies pinned here in their project.
        #
        # Generally packages listed here are pinned to a major version range.
        #
        # e.g.
        # Python FooBar package for foobaring
        # pyfoobar>=1.0, <2.0
        #
        # This will allow for any consuming projects to use this library as
        # long as they have a version of pyfoobar equal to or greater than 1.x
        # and less than 2.x installed.
    ],
    extras_require={
        ':python_version=="2.7"': ['typing>=3.6'],  # allow typehinting PY2
    },
)

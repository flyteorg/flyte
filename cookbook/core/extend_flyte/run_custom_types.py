"""
.. _advanced_custom_types:

Writing Custom Types
--------------------

Flyte is a strongly typed framework for authoring tasks and workflows. But, there are situations when the existing set
of types do not directly work. This is true with any programming language. This is when the languages support higher
level concepts to describe User specific objects - like classes in python/java/c++, struct in C/golang, etc

Flytekit allows modeling user classes similarly. The idea is to make an interface that is more productive for the
usecase, but write a transformer that transforms the user defined type to one of the generic constructs in Flyte's
Type system.

In this example, we will try to model an example user defined set and show how it can be integrated seamlessly with
Flytekit's typing engine.


"""
import os
import tempfile
import typing

# %%
# FlyteContext is used only to access a random local directory
from typing import Type

# %%
# Defined type here represents a list of Files on the disk. We will refer to it as ``MyDataset``
from flytekit import FlyteContext, task, workflow
from flytekit.extend import TypeEngine, TypeTransformer
from flytekit.models.core.types import BlobType
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType


class MyDataset(object):
    """
    Dataset here is a set of files that exist together. In Flyte this maps to a Multi-part blob or a directory
    """

    def __init__(self, base_dir: str = None):
        if base_dir is None:
            self._tmp_dir = tempfile.TemporaryDirectory()
            self._base_dir = self._tmp_dir.name
            self._files = []
        else:
            self._base_dir = base_dir
            files = os.listdir(base_dir)
            self._files = [os.path.join(base_dir, f) for f in files]

    @property
    def base_dir(self) -> str:
        return self._base_dir

    @property
    def files(self) -> typing.List[str]:
        return self._files

    def new_file(self, name: str) -> str:
        new_file = os.path.join(self._base_dir, name)
        self._files.append(new_file)
        return new_file


# %%
# ``MyDataset`` represents a set of files locally, but, when a workflow consists of multiple steps, we want the data to
# flow between the different steps. To achieve this, it is necessary to explain how the data will be transformed to
# Flyte's remote references. To do this, we create a new instance of
# :py:class:`flytekit.core.type_engine.TypeTransformer`, for the type ``MyDataset`` as follows
#
# .. note::
#
#   The TypeTransformer is a Generic abstract base class. The Generic type argument here refers to the actual object
#   that we want to work with. In this case, it is the ``MyDataset`` object
class MyDatasetTransformer(TypeTransformer[MyDataset]):
    _TYPE_INFO = BlobType(
        format="binary", dimensionality=BlobType.BlobDimensionality.MULTIPART
    )

    def __init__(self):
        super(MyDatasetTransformer, self).__init__(
            name="mydataset-transform", t=MyDataset
        )

    def get_literal_type(self, t: Type[MyDataset]) -> LiteralType:
        """
        This is useful to tell the Flytekit type system that ``MyDataset`` actually refers to what corresponding type
        In this example, we say its of format binary (do not try to introspect) and there are more than one files in it
        """
        return LiteralType(blob=self._TYPE_INFO)

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: MyDataset,
        python_type: Type[MyDataset],
        expected: LiteralType,
    ) -> Literal:
        """
        This method is used to convert from given python type object ``MyDataset`` to the Literal representation
        """
        # Step 1: lets upload all the data into a remote place recommended by Flyte
        remote_dir = ctx.file_access.get_random_remote_directory()
        ctx.file_access.upload_directory(python_val.base_dir, remote_dir)
        # Step 2: lets return a pointer to this remote_dir in the form of a literal
        return Literal(
            scalar=Scalar(
                blob=Blob(uri=remote_dir, metadata=BlobMetadata(type=self._TYPE_INFO))
            )
        )

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[MyDataset]
    ) -> MyDataset:
        """
        In this function we want to be able to re-hydrate the custom object from Flyte Literal value
        """
        # Step 1: lets download remote data locally
        local_dir = ctx.file_access.get_random_local_directory()
        ctx.file_access.download_directory(lv.scalar.blob.uri, local_dir)
        # Step 2: create the MyDataset object
        return MyDataset(base_dir=local_dir)


# %%
# Before we can use MyDataset in our tasks, we need to let flytekit know that ``MyDataset`` should be considered as a
# valid type. This is done using the :py:func:`flytekit.core.type_engine.TypeEngine.register` function.
TypeEngine.register(MyDatasetTransformer())


# %%
# Now the new type should be ready to use. Let us write an example generator and consumer for this new datatype


@task
def generate() -> MyDataset:
    d = MyDataset()
    for i in range(3):
        fp = d.new_file(f"x{i}")
        with open(fp, "w") as f:
            f.write(f"Contents of file{i}")

    return d


@task
def consume(d: MyDataset) -> str:
    s = ""
    for f in d.files:
        with open(f) as fp:
            s += fp.read()
            s += "\n"
    return s


@workflow
def wf() -> str:
    return consume(d=generate())


# %%
# We can run this workflow locally and test it. Remember even when you run it locally, flytekit will excercise the
# entire path

if __name__ == "__main__":
    print(wf())

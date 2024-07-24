import os
import pathlib
import typing

from google.protobuf.json_format import MessageToJson
from rich import print

from flytekit import BlobType, Literal
from flytekit.core.data_persistence import FileAccessProvider
from flytekit.interaction.rich_utils import RichCallback
from flytekit.interaction.string_literals import literal_string_repr


def download_literal(
    file_access: FileAccessProvider, var: str, data: Literal, download_to: typing.Optional[pathlib.Path] = None
):
    """
    Download a single literal to a file, if it is a blob or structured dataset.
    """
    if data is None:
        print(f"Skipping {var} as it is None.")
        return
    if data.scalar:
        if data.scalar and (data.scalar.blob or data.scalar.structured_dataset):
            uri = data.scalar.blob.uri if data.scalar.blob else data.scalar.structured_dataset.uri
            if uri is None:
                print("No data to download.")
                return
            is_multipart = False
            if data.scalar.blob:
                is_multipart = data.scalar.blob.metadata.type.dimensionality == BlobType.BlobDimensionality.MULTIPART
            elif data.scalar.structured_dataset:
                is_multipart = True
            file_access.get_data(
                uri, str(download_to / var) + os.sep, is_multipart=is_multipart, callback=RichCallback()
            )
        elif data.scalar.union is not None:
            download_literal(file_access, var, data.scalar.union.value, download_to)
        elif data.scalar.generic is not None:
            with open(download_to / f"{var}.json", "w") as f:
                f.write(MessageToJson(data.scalar.generic))
        else:
            print(
                f"[dim]Skipping {var} val {literal_string_repr(data)} as it is not a blob, structured dataset,"
                f" or generic type.[/dim]"
            )
            return
    elif data.collection:
        for i, v in enumerate(data.collection.literals):
            download_literal(file_access, f"{i}", v, download_to / var)
    elif data.map:
        download_to = pathlib.Path(download_to)
        for k, v in data.map.literals.items():
            download_literal(file_access, f"{k}", v, download_to / var)
    print(f"Downloaded f{var} to {download_to}")

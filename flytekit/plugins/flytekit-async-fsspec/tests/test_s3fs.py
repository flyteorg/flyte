import os
from unittest import mock
from unittest.mock import MagicMock, mock_open

import pytest
from flytekitplugins.async_fsspec import AsyncS3FileSystem
from flytekitplugins.async_fsspec.s3fs.constants import DEFAULT_DOWNLOAD_CHUNK_SIZE, DEFAULT_UPLOAD_CHUNK_SIZE


@mock.patch("flytekitplugins.async_fsspec.AsyncS3FileSystem._parent")
@mock.patch("flytekitplugins.async_fsspec.AsyncS3FileSystem.invalidate_cache")
@mock.patch("flytekitplugins.async_fsspec.AsyncS3FileSystem._call_s3")
@mock.patch("mimetypes.guess_type")
@mock.patch("os.path.getsize")
@pytest.mark.asyncio
async def test_put_file_single_object_upload(
    mock_getsize, mock_guess_type, mock_call_s3, mock_invalidate_cache, mock_parent
):
    mock_bucket = "mock-bucket"
    mock_file_name = "mock_file_name"
    mock_file_size = 32 * 2**20  # 32MB
    mock_getsize.return_value = mock_file_size
    mock_guess_type.return_value = (None, None)
    mock_parent.return_value = None
    mock_body = os.urandom(mock_file_size)
    m = mock_open(read_data=mock_body)

    with mock.patch("builtins.open", m):
        asyncs3fs = AsyncS3FileSystem()
        await asyncs3fs._put_file(lpath=f"/{mock_file_name}", rpath=f"s3://{mock_bucket}/{mock_file_name}")

    mock_call_s3.assert_called_once_with("put_object", Bucket=mock_bucket, Key=mock_file_name, Body=mock_body)


@mock.patch("flytekitplugins.async_fsspec.AsyncS3FileSystem._parent")
@mock.patch("flytekitplugins.async_fsspec.AsyncS3FileSystem.invalidate_cache")
@mock.patch("flytekitplugins.async_fsspec.AsyncS3FileSystem._call_s3")
@mock.patch("mimetypes.guess_type")
@mock.patch("os.path.getsize")
@pytest.mark.asyncio
async def test_put_file_multipart_upload(
    mock_getsize, mock_guess_type, mock_call_s3, mock_invalidate_cache, mock_parent
):
    mock_bucket = "mock-bucket"
    mock_file_name = "mock_file_name"
    mock_upload_id = "mock_upload_id"
    mock_ETag = "mock_ETag"
    mock_file_size = 256 * 2**20  # 256MB
    mock_getsize.return_value = mock_file_size
    mock_guess_type.return_value = (None, None)

    def call_s3_side_effect(*args, **kwargs):
        if args[0] == "create_multipart_upload":
            return {"UploadId": mock_upload_id}
        elif args[0] == "upload_part":
            part_number = kwargs["PartNumber"]
            return {"ETag": f"{mock_ETag}{part_number}"}
        elif args[0] == "complete_multipart_upload":
            return None

    mock_call_s3.side_effect = call_s3_side_effect

    mock_parent.return_value = None

    mock_body = os.urandom(mock_file_size)
    m = mock_open(read_data=mock_body)

    with mock.patch("builtins.open", m):
        asyncs3fs = AsyncS3FileSystem()
        await asyncs3fs._put_file(lpath=f"/{mock_file_name}", rpath=f"s3://{mock_bucket}/{mock_file_name}")

    mock_chunk_count = mock_file_size // DEFAULT_UPLOAD_CHUNK_SIZE
    if mock_file_size % DEFAULT_UPLOAD_CHUNK_SIZE > 0:
        mock_chunk_count += 1
    put_object_calls = []
    for i in range(mock_chunk_count):
        part_number = i + 1
        start_byte = i * DEFAULT_UPLOAD_CHUNK_SIZE
        end_byte = min(start_byte + DEFAULT_UPLOAD_CHUNK_SIZE, mock_file_size)
        body = mock_body[start_byte:end_byte]
        put_object_calls.append(
            mock.call(
                "upload_part",
                Bucket=mock_bucket,
                Key=mock_file_name,
                PartNumber=part_number,
                UploadId=mock_upload_id,
                Body=body,
            ),
        )

    mock_call_s3.assert_has_calls(
        put_object_calls
        + [
            mock.call("create_multipart_upload", Bucket=mock_bucket, Key=mock_file_name),
            mock.call(
                "complete_multipart_upload",
                Bucket=mock_bucket,
                Key=mock_file_name,
                UploadId=mock_upload_id,
                MultipartUpload={
                    "Parts": [{"PartNumber": i, "ETag": f"{mock_ETag}{i}"} for i in range(1, mock_chunk_count + 1)]
                },
            ),
        ],
        any_order=True,
    )
    assert mock_call_s3.call_count == 2 + mock_chunk_count


@mock.patch("flytekitplugins.async_fsspec.AsyncS3FileSystem._call_s3")
@mock.patch("flytekitplugins.async_fsspec.AsyncS3FileSystem._info")
@mock.patch("os.path.isdir")
@pytest.mark.asyncio
async def test_get_file_file_size_is_none(mock_isdir, mock_info, mock_call_s3):
    mock_bucket = "mock-bucket"
    mock_file_name = "mock_file_name"
    mock_file_size = 32 * 2**20  # 32MB
    mock_isdir.return_value = False
    mock_info.return_value = {"size": None}

    file_been_read = 0

    async def read_side_effect(*args, **kwargs):
        read_size = args[0]
        nonlocal file_been_read
        real_read_size = min(read_size, mock_file_size - file_been_read)
        if real_read_size == 0:
            return None
        file_been_read += real_read_size
        return os.urandom(real_read_size)

    mock_chunk = MagicMock()
    mock_chunk.read.side_effect = read_side_effect
    mock_call_s3.return_value = {"Body": mock_chunk, "ContentLength": mock_file_size}

    m = mock_open()

    with mock.patch("builtins.open", m):
        asyncs3fs = AsyncS3FileSystem()
        await asyncs3fs._get_file(lpath=f"/{mock_file_name}", rpath=f"s3://{mock_bucket}/{mock_file_name}")

    assert file_been_read == mock_file_size
    mock_call_s3.assert_called_once_with("get_object", Bucket=mock_bucket, Key=mock_file_name, Range="bytes=0")


@mock.patch("flytekitplugins.async_fsspec.AsyncS3FileSystem._call_s3")
@mock.patch("flytekitplugins.async_fsspec.AsyncS3FileSystem._info")
@mock.patch("os.path.isdir")
@pytest.mark.asyncio
async def test_get_file_file_size_is_not_none(mock_isdir, mock_info, mock_call_s3):
    mock_bucket = "mock-bucket"
    mock_file_name = "mock_file_name"
    mock_file_size = 256 * 2**20  # 256MB
    mock_isdir.return_value = False
    mock_info.return_value = {"size": mock_file_size}

    file_been_read = 0

    def call_s3_side_effect(*args, **kwargs):
        start_byte, end_byte = kwargs["Range"][6:].split("-")
        start_byte, end_byte = int(start_byte), int(end_byte)
        chunk_size = end_byte - start_byte + 1
        chunk_been_read = 0

        async def read_side_effect(*args, **kwargs):
            nonlocal chunk_been_read
            nonlocal file_been_read
            read_size = args[0]
            real_read_size = min(read_size, chunk_size - chunk_been_read)
            if real_read_size == 0:
                return None
            chunk_been_read += real_read_size
            file_been_read += real_read_size
            return os.urandom(real_read_size)

        mock_chunk = MagicMock()
        mock_chunk.read.side_effect = read_side_effect
        return {"Body": mock_chunk, "ContentLength": chunk_size}

    mock_call_s3.side_effect = call_s3_side_effect

    m = mock_open()
    with mock.patch("builtins.open", m):
        asyncs3fs = AsyncS3FileSystem()
        await asyncs3fs._get_file(lpath=f"/{mock_file_name}", rpath=f"s3://{mock_bucket}/{mock_file_name}")

    assert file_been_read == mock_file_size

    mock_chunk_count = mock_file_size // DEFAULT_DOWNLOAD_CHUNK_SIZE
    if mock_file_size % DEFAULT_DOWNLOAD_CHUNK_SIZE > 0:
        mock_chunk_count += 1
    get_object_calls = []
    for i in range(mock_chunk_count):
        start_byte = i * DEFAULT_DOWNLOAD_CHUNK_SIZE
        end_byte = min(start_byte + DEFAULT_DOWNLOAD_CHUNK_SIZE, mock_file_size) - 1
        get_object_calls.append(
            mock.call("get_object", Bucket=mock_bucket, Key=mock_file_name, Range=f"bytes={start_byte}-{end_byte}")
        )
    mock_call_s3.assert_has_calls(get_object_calls)
    assert mock_call_s3.call_count == len(get_object_calls)

import asyncio
import mimetypes
import os

from fsspec.callbacks import _DEFAULT_CALLBACK
from s3fs import S3FileSystem
from s3fs.core import S3_RETRYABLE_ERRORS, version_id_kw

from .constants import (
    DEFAULT_CONCURRENT_DOWNLOAD,
    DEFAULT_CONCURRENT_UPLOAD,
    DEFAULT_DOWNLOAD_BODY_READ_SIZE,
    DEFAULT_DOWNLOAD_CHUNK_SIZE,
    DEFAULT_UPLOAD_CHUNK_SIZE,
    SINGLE_OBJECT_UPLOAD_LIMIT,
)


class AsyncS3FileSystem(S3FileSystem):
    def __init__(self, **s3kwargs):
        super().__init__(**s3kwargs)

    async def _put_file(
        self,
        lpath,
        rpath,
        callback=_DEFAULT_CALLBACK,
        chunksize=DEFAULT_UPLOAD_CHUNK_SIZE,
        concurrent_upload=DEFAULT_CONCURRENT_UPLOAD,
        **kwargs,
    ):
        """
        Put a file from lpath to rpath.
        Args:
            lpath (str): The local path of the file to be uploaded.
            rpath (str): The remote path which the file should be uploaded to.
            callback (function, optional): The callback function.
            chunksize (int, optional): Upload chunksize. Defaults to 50 * 2**20 (50MB).
            concurrent_upload (int, optional): The number of concurrent upload when using multipart upload. Defaults to 4.
        """
        bucket, key, _ = self.split_path(rpath)
        if os.path.isdir(lpath):
            if key:
                # don't make remote "directory"
                return
            else:
                await self._mkdir(lpath)
        size = os.path.getsize(lpath)
        callback.set_size(size)

        if "ContentType" not in kwargs:
            content_type, _ = mimetypes.guess_type(lpath)
            if content_type is not None:
                kwargs["ContentType"] = content_type

        with open(lpath, "rb") as f0:
            if size < min(SINGLE_OBJECT_UPLOAD_LIMIT, 2 * chunksize):
                chunk = f0.read()
                await self._call_s3("put_object", Bucket=bucket, Key=key, Body=chunk, **kwargs)
                callback.relative_update(size)
            else:
                mpu = await self._call_s3("create_multipart_upload", Bucket=bucket, Key=key, **kwargs)

                # async function to upload a single chunk
                async def upload_chunk(chunk, part_number):
                    result = await self._call_s3(
                        "upload_part",
                        Bucket=bucket,
                        PartNumber=part_number,
                        UploadId=mpu["UploadId"],
                        Body=chunk,
                        Key=key,
                    )
                    callback.relative_update(len(chunk))
                    return {"PartNumber": part_number, "ETag": result["ETag"]}

                tasks = set()
                part_number = 1
                parts = []
                read_end = False
                while True:
                    while len(tasks) < concurrent_upload:
                        chunk = f0.read(chunksize)
                        if not chunk:
                            read_end = True
                            break
                        tasks.add(asyncio.create_task(upload_chunk(chunk, part_number)))
                        part_number += 1
                    if read_end:
                        break
                    done, pending = await asyncio.wait(tasks, return_when=asyncio.FIRST_COMPLETED)
                    parts.extend(map(lambda x: x.result(), done))
                    tasks = pending

                parts.extend(await asyncio.gather(*tasks))
                parts.sort(key=lambda part: part["PartNumber"])
                await self._call_s3(
                    "complete_multipart_upload",
                    Bucket=bucket,
                    Key=key,
                    UploadId=mpu["UploadId"],
                    MultipartUpload={"Parts": parts},
                )
        while rpath:
            self.invalidate_cache(rpath)
            rpath = self._parent(rpath)

    async def _get_file(
        self,
        rpath,
        lpath,
        callback=_DEFAULT_CALLBACK,
        version_id=None,
        chunksize=DEFAULT_DOWNLOAD_CHUNK_SIZE,
        concurrent_download=DEFAULT_CONCURRENT_DOWNLOAD,
    ):
        """
        Get a file from rpath to lpath.
        Args:
            rpath (str): The remote path of the file to be downloaded.
            lpath (str): The local path which the file should be downloaded to.
            callback (function, optional): The callback function.
            chunksize (int, optional): Download chunksize. Defaults to 50 * 2**20 (50MB).
            version_id (str, optional): The version id of the file. Defaults to None.
        """
        if os.path.isdir(lpath):
            return

        # get the file size
        file_info = await self._info(path=rpath, version_id=version_id)
        file_size = file_info["size"]

        bucket, key, vers = self.split_path(rpath)

        # the async function to get a range of the remote file
        async def _open_file(start_byte: int, end_byte: int = None):
            kw = self.req_kw.copy()
            if end_byte:
                kw["Range"] = f"bytes={start_byte}-{end_byte}"
            else:
                kw["Range"] = f"bytes={start_byte}"
            resp = await self._call_s3(
                "get_object",
                Bucket=bucket,
                Key=key,
                **version_id_kw(version_id or vers),
                **kw,
            )
            return resp["Body"], resp.get("ContentLength", None)

        # Refer to s3fs's implementation
        async def handle_read_error(body, failed_reads, restart_byte, end_byte=None):
            if failed_reads >= self.retries:
                raise
            try:
                body.close()
            except Exception:
                pass

            await asyncio.sleep(min(1.7**failed_reads * 0.1, 15))
            body, _ = await _open_file(restart_byte, end_byte)
            return body

        # According to s3fs documentation, some file systems might not be able to measure the file’s size,
        # in which case, the returned dict will include 'size': None. When we cannot get the file size
        # in advance, we keep using the original implementation of s3fs.
        if file_size is None:
            # From s3fs
            body, content_length = await _open_file(start_byte=0)
            callback.set_size(content_length)

            failed_reads = 0
            bytes_read = 0

            try:
                with open(lpath, "wb") as f0:
                    while True:
                        try:
                            chunk = await body.read(DEFAULT_DOWNLOAD_BODY_READ_SIZE)
                        except S3_RETRYABLE_ERRORS:
                            failed_reads += 1
                            body = await handle_read_error(body, failed_reads, bytes_read)
                            continue

                        if not chunk:
                            break

                        f0.write(chunk)
                        bytes_read += len(chunk)
                        callback.relative_update(len(chunk))
            finally:
                try:
                    body.close()
                except Exception:
                    pass
        else:
            callback.set_size(file_size)
            with open(lpath, "wb") as f0:
                # async function to download a single chunk
                async def download_chunk(chunk_index: int):
                    start_byte = chunk_index * chunksize
                    end_byte = min(start_byte + chunksize, file_size) - 1
                    body, _ = await _open_file(start_byte, end_byte)
                    failed_reads = 0
                    bytes_read = 0
                    try:
                        while True:
                            try:
                                chunk = await body.read(DEFAULT_DOWNLOAD_BODY_READ_SIZE)
                            except S3_RETRYABLE_ERRORS:
                                failed_reads += 1
                                body = await handle_read_error(body, failed_reads, start_byte + bytes_read, end_byte)
                                continue

                            if not chunk:
                                break

                            f0.seek(start_byte + bytes_read)
                            f0.write(chunk)
                            bytes_read += len(chunk)
                            callback.relative_update(len(chunk))
                    finally:
                        try:
                            body.close()
                        except Exception:
                            pass

                chunk_count = file_size // chunksize
                if file_size % chunksize > 0:
                    chunk_count += 1

                tasks = set()
                current_chunk = 0
                while current_chunk < chunk_count:
                    while current_chunk < chunk_count and len(tasks) < concurrent_download:
                        tasks.add(asyncio.create_task(download_chunk(current_chunk)))
                        current_chunk += 1
                    _, pending = await asyncio.wait(tasks, return_when=asyncio.FIRST_COMPLETED)
                    tasks = pending
                await asyncio.gather(*tasks)

import json
import shlex
import subprocess
from dataclasses import asdict, dataclass
from tempfile import NamedTemporaryFile
from typing import Optional

from flyteidl.admin.agent_pb2 import CreateTaskResponse, DeleteTaskResponse, GetTaskResponse, Resource
from flytekitplugins.mmcloud.utils import async_check_output, mmcloud_status_to_flyte_phase

from flytekit import current_context
from flytekit.extend.backend.base_agent import AgentBase, AgentRegistry
from flytekit.loggers import logger
from flytekit.models.literals import LiteralMap
from flytekit.models.task import TaskTemplate


@dataclass
class Metadata:
    job_id: str


class MMCloudAgent(AgentBase):
    name = "MMCloud Agent"

    def __init__(self):
        super().__init__(task_type="mmcloud_task", asynchronous=True)
        self._response_format = ["--format", "json"]

    async def async_login(self):
        """
        Log in to Memory Machine Cloud OpCenter.
        """
        try:
            # If already logged in, this will reset the session timer
            login_info_command = ["float", "login", "--info"]
            await async_check_output(*login_info_command)
        except subprocess.CalledProcessError:
            logger.info("Attempting to log in to OpCenter")
            try:
                secrets = current_context().secrets
                login_command = [
                    "float",
                    "login",
                    "--address",
                    secrets.get("mmc_address"),
                    "--username",
                    secrets.get("mmc_username"),
                    "--password",
                    secrets.get("mmc_password"),
                ]
                await async_check_output(*login_command)
            except subprocess.CalledProcessError:
                logger.exception("Failed to log in to OpCenter")
                raise

            logger.info("Logged in to OpCenter")

    async def create(
        self, output_prefix: str, task_template: TaskTemplate, inputs: Optional[LiteralMap] = None, **kwargs
    ) -> CreateTaskResponse:
        """
        Submit Flyte task as MMCloud job to the OpCenter, and return the job UID for the task.
        """
        submit_command = [
            "float",
            "submit",
            "--force",
            *self._response_format,
        ]

        # We do not use container.resources because FlytePropeller will impose limits that should not apply to MMCloud
        min_cpu, min_mem, max_cpu, max_mem = task_template.custom["resources"]
        submit_command.extend(["--cpu", f"{min_cpu}:{max_cpu}"] if max_cpu else ["--cpu", f"{min_cpu}"])
        submit_command.extend(["--mem", f"{min_mem}:{max_mem}"] if max_mem else ["--mem", f"{min_mem}"])

        container = task_template.container

        image = container.image
        submit_command.extend(["--image", image])

        env = container.env
        for key, value in env.items():
            submit_command.extend(["--env", f"{key}={value}"])

        submit_extra = task_template.custom["submit_extra"]
        submit_command.extend(shlex.split(submit_extra))

        args = task_template.container.args
        script_lines = ["#!/bin/bash\n", f"{shlex.join(args)}\n"]

        task_id = task_template.id
        try:
            # float binary takes a job file as input, so one must be created
            # Use a uniquely named temporary file to avoid race conditions and clutter
            with NamedTemporaryFile(mode="w") as job_file:
                job_file.writelines(script_lines)
                # Flush immediately so that the job file is usable
                job_file.flush()
                logger.debug("Wrote job script")

                submit_command.extend(["--job", job_file.name])

                logger.info(f"Attempting to submit Flyte task {task_id} as MMCloud job")
                logger.debug(f"With command: {submit_command}")
                try:
                    await self.async_login()
                    submit_response = await async_check_output(*submit_command)
                    submit_response = json.loads(submit_response.decode())
                    job_id = submit_response["id"]
                except subprocess.CalledProcessError as e:
                    logger.exception(
                        f"Failed to submit Flyte task {task_id} as MMCloud job\n"
                        f"[stdout] {e.stdout.decode()}\n"
                        f"[stderr] {e.stderr.decode()}\n"
                    )
                    raise
                except (UnicodeError, json.JSONDecodeError):
                    logger.exception(f"Failed to decode submit response for Flyte task: {task_id}")
                    raise
                except KeyError:
                    logger.exception(f"Failed to obtain MMCloud job id for Flyte task: {task_id}")
                    raise

                logger.info(f"Submitted Flyte task {task_id} as MMCloud job {job_id}")
                logger.debug(f"OpCenter response: {submit_response}")
        except OSError:
            logger.exception("Cannot open job script for writing")
            raise

        metadata = Metadata(job_id=job_id)

        return CreateTaskResponse(resource_meta=json.dumps(asdict(metadata)).encode("utf-8"))

    async def async_get(self, resource_meta: bytes, **kwargs) -> GetTaskResponse:
        """
        Return the status of the task, and return the outputs on success.
        """
        metadata = Metadata(**json.loads(resource_meta.decode("utf-8")))
        job_id = metadata.job_id

        show_command = [
            "float",
            "show",
            *self._response_format,
            "--job",
            job_id,
        ]

        logger.info(f"Attempting to obtain status for MMCloud job {job_id}")
        logger.debug(f"With command: {show_command}")
        try:
            await self.async_login()
            show_response = await async_check_output(*show_command)
            show_response = json.loads(show_response.decode())
            job_status = show_response["status"]
        except subprocess.CalledProcessError as e:
            logger.exception(
                f"Failed to get show response for MMCloud job: {job_id}\n"
                f"[stdout] {e.stdout.decode()}\n"
                f"[stderr] {e.stderr.decode()}\n"
            )
            raise
        except (UnicodeError, json.JSONDecodeError):
            logger.exception(f"Failed to decode show response for MMCloud job: {job_id}")
            raise
        except KeyError:
            logger.exception(f"Failed to obtain status for MMCloud job: {job_id}")
            raise

        task_phase = mmcloud_status_to_flyte_phase(job_status)

        logger.info(f"Obtained status for MMCloud job {job_id}: {job_status}")
        logger.debug(f"OpCenter response: {show_response}")

        return GetTaskResponse(resource=Resource(phase=task_phase))

    async def async_delete(self, resource_meta: bytes, **kwargs) -> DeleteTaskResponse:
        """
        Delete the task. This call should be idempotent.
        """
        metadata = Metadata(**json.loads(resource_meta.decode("utf-8")))
        job_id = metadata.job_id

        cancel_command = [
            "float",
            "cancel",
            "--force",
            "--job",
            job_id,
        ]

        logger.info(f"Attempting to cancel MMCloud job {job_id}")
        logger.debug(f"With command: {cancel_command}")
        try:
            await self.async_login()
            await async_check_output(*cancel_command)
        except subprocess.CalledProcessError as e:
            logger.exception(
                f"Failed to cancel MMCloud job: {job_id}\n[stdout] {e.stdout.decode()}\n[stderr] {e.stderr.decode()}\n"
            )
            raise

        logger.info(f"Submitted cancel request for MMCloud job: {job_id}")

        return DeleteTaskResponse()


AgentRegistry.register(MMCloudAgent())

import os
from enum import Enum


from union.actor import ActorEnvironment

from dataclasses import dataclass


class ServerlessEnvironmentType(str, Enum):
    STAGING = "STAGING"


@dataclass
class EnvironmentConfig:
    type: ServerlessEnvironmentType = ServerlessEnvironmentType.STAGING

    class Config:
        json_encoders = {ServerlessEnvironmentType: lambda v: v.value}


StagingEnvironment = EnvironmentConfig(type=ServerlessEnvironmentType.STAGING)


actor = ActorEnvironment(
    name="xs",
    container_image=os.getenv("UNION_RUNTIME_TEST_IMAGE"),
    replica_count=1,
)


@dataclass
class User:
    name: str = None


@actor.task
def resolve_environment_config(env: str) -> EnvironmentConfig:
    return StagingEnvironment

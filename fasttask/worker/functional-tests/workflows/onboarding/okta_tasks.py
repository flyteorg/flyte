from typing import List

from env import EnvironmentConfig, User, actor
from flytekit import workflow


@actor.task
def get_approved_users(users: List[User]) -> List[User]:
    return users


@actor.task
def get_pending_users(env_config: EnvironmentConfig, is_workshop: bool) -> List[User]:
    import os

    return [User(name=os.getcwd())]


@workflow
def get_segregated_pending_users(env_config: EnvironmentConfig, is_workshop: bool) -> List[User]:
    users = get_pending_users(env_config=env_config, is_workshop=is_workshop)
    return get_approved_users(users)

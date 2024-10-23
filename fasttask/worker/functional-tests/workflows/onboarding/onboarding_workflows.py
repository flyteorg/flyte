from flytekit import workflow


from env import resolve_environment_config
import okta_tasks


@workflow
def approve_pending_users(env: str = "staging"):
    env_config = resolve_environment_config(env=env)
    okta_tasks.get_segregated_pending_users(env_config=env_config, is_workshop=False)

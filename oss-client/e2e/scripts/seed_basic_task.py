import asyncio
import os

import flyte
from write_run_metadata import write_run_metadata

env_name ="task_env"
logs = "This is a task log"
input = 100

env = flyte.TaskEnvironment(name=env_name)

@env.task
async def add_ten(x: int) -> int:
  print(logs)
  return x + 10


if __name__ == "__main__":
    api_key = os.getenv('UNION_API_KEY')
    if api_key is None:
      raise Exception('UNION_API_KEY has not been provided')
    # These values (org, project, domain) can be updated in buildkite
    org = os.getenv('TEST_ORG', 'playground')
    test_project = os.getenv('TEST_PROJECT', 'e2e')
    test_domain = os.getenv('TEST_DOMAIN', 'development')

    flyte.init(api_key=api_key, org=org, project=test_project, domain=test_domain, image_builder="remote")
    deployments = flyte.deploy(env)

    run = flyte.with_runcontext().run(add_ten, x=input)
    print(run.name)
    print(run.url)
    write_run_metadata({
        "env": env_name,
        "name": "add_ten",
        "id": run.name,
        "url": run.url,
        "display_name": "add_ten",
        "inputs": input,
        "logs": logs
    })

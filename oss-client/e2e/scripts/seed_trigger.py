import os
import flyte


env_name ="trigger_env"
logs = "This is a trigger log"
interval_minutes=15

env = flyte.TaskEnvironment(name=env_name)

@env.task(triggers=flyte.Trigger(
   name="e2e-trigger-list",
   description="An excellent trigger for tasks",
   automation=flyte.FixedRate(interval_minutes=interval_minutes,start_time=None)
))
def list_trigger() -> str:
    return 'trigger'

if __name__ == "__main__":
    api_key = os.getenv('UNION_API_KEY')
    if api_key is None:
      raise Exception('UNION_API_KEY has not been provided')
    org = os.getenv('TEST_ORG', 'playground')
    test_project = os.getenv('TEST_PROJECT', 'e2e')
    test_domain = os.getenv('TEST_DOMAIN', 'development')

    flyte.init(api_key=api_key, org=org, project=test_project, domain=test_domain, image_builder="remote")
    deployments = flyte.deploy(env)

    from write_run_metadata import write_run_metadata
    write_run_metadata({
        "env": env_name,
        "name": "list_trigger",
        "interval": interval_minutes,
    })



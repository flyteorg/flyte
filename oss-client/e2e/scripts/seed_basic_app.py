"""A simple "Hello World" FastAPI app example for serving."""

import os
import flyte
from flyte.remote import App
from flyte.app.extras import FastAPIAppEnvironment
from fastapi import FastAPI

# Define a simple FastAPI application
app = FastAPI(
    title="Hello World API",
    description="A simple FastAPI application",
    version="1.0.0",
)

# Create an AppEnvironment for the FastAPI app
env = FastAPIAppEnvironment(
    name="hello-app-3", # renamed to create new app because of bad data in test environment
    app=app,
    image=flyte.Image.from_debian_base(python_version=(3, 12)).with_pip_packages(
        "fastapi",
        "uvicorn",
    ),
    resources=flyte.Resources(cpu=1, memory="512Mi"),
    requires_auth=False,
)

# Define API endpoints
@app.get("/")
async def root():
    return {"message": "Hello, World!"}

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

# Running this script will deploy and serve the app on your Union/Flyte instance.
if __name__ == "__main__":
    # Initialize Flyte using API key and environment configuration.
    from write_run_metadata import write_run_metadata
    api_key = os.getenv('UNION_API_KEY')
    if api_key is None:
        raise Exception('UNION_API_KEY has not been provided')
    # These values (org, project, domain) can be updated in buildkite
    org = os.getenv('TEST_ORG', 'playground')
    test_project = os.getenv('TEST_PROJECT', 'e2e')
    test_domain = os.getenv('TEST_DOMAIN', 'development')

    flyte.init(api_key=api_key, org=org, project=test_project, domain=test_domain, image_builder="remote")
    print('about to deploy the app')
    # Deploy the app remotely.
    flyte.deploy(env)
    print('just deployed. about to activate the app')
    # Activate the app
    app = App.get(name=env.name)
    app.activate()

    # Print the app URL.
    print(app.url)
    print("App 'hello-app' is now deployed.")
    write_run_metadata({
        "env": "hello-app-3",
        "name": "hello_app",
        "id": app.name,
        "url": app.url,
        "display_name": "hello_app",
        "inputs": None,
        "logs": f"App URL: {app.url}"
    })
